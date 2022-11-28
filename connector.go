package connector

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/nakji-network/connector/chain"
	"github.com/nakji-network/connector/config"
	"github.com/nakji-network/connector/kafkautils"
	"github.com/nakji-network/connector/monitor"
	"github.com/nakji-network/connector/protoregistry"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/heptiolabs/healthcheck"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type Connector struct {
	*kafkautils.Consumer
	*kafkautils.Producer

	consumerStarted   bool
	env               kafkautils.Env
	kafkaUrl          string
	manifest          *manifest
	producerStarted   bool
	protoRegistryHost string

	RPCMap map[string]chain.RPCs
	Config *viper.Viper
	Health healthcheck.Handler

	//	DEPRECATED, please use ProduceWithTransaction instead
	//	EventSink can be used to push incoming on-chain events to Kafka.
	// 	All kafka Produce logic will be handled under the hood.
	EventSink chan<- *kafkautils.Message
}

type Option func(*Connector)

// NewConnector returns a base connector implementation that other connectors can embed to add on to.
func NewConnector(options ...Option) (*Connector, error) {
	conf := config.GetConfig()
	conf.SetDefault("kafka.env", "dev")
	conf.SetDefault("protoregistry.host", "localhost:9191")
	conf.SetDefault("trace.sample.ratio", 0.2)
	conf.SetDefault("trace.host.grpc", "localhost:4317")
	conf.SetDefault("trace.host.timeout", "2s")

	rpcMap := make(map[string]chain.RPCs)
	err := conf.UnmarshalKey("rpcs", &rpcMap)
	if err != nil {
		return nil, fmt.Errorf("could not load RPC list from config file")
	}

	c := &Connector{
		manifest:          LoadManifest(),
		env:               kafkautils.Env(conf.GetString("kafka.env")),
		kafkaUrl:          conf.GetString("kafka.url"),
		Health:            healthcheck.NewHandler(),
		RPCMap:            rpcMap,
		protoRegistryHost: conf.GetString("protoregistry.host"),
	}

	parseOptions(c, options...)

	if c.manifest == nil {
		log.Fatal().Msg("missing manifest.yaml")
	}

	// allow users to access configs in the namespace `c.id()`
	c.Config = conf.Sub(c.id())
	if c.Config == nil {
		c.Config = viper.New()
	}

	log.Info().
		Str("id", c.id()).
		Msg("Starting connector")

	//	Create kafka Produce channel and provide an outlet to connector object
	eventSink := make(chan *kafkautils.Message, 10000)
	go c.initProduceChannel(eventSink)
	c.EventSink = eventSink

	// For Liveness and Readiness Probe checks
	go http.ListenAndServe("0.0.0.0:8080", c.Health)
	log.Info().Str("addr", "0.0.0.0:8080").Msg("healthcheck listening on /live and /ready")

	//	Initialize trace provider
	ctx, cancel := context.WithTimeout(context.TODO(), conf.GetDuration("trace.host.timeout"))
	defer cancel()
	_, err = monitor.InitTracerProvider(ctx, conf.GetString("trace.host.grpc"), c.id(), c.manifest.Version.String(), string(c.env), conf.GetFloat64("trace.sample.ratio"))
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize trace provider")
	}

	//	Start Prometheus monitoring
	monitor.StartMonitor(c.id())

	return c, nil
}

func parseOptions(c *Connector, options ...Option) {
	for _, option := range options {
		option(c)
	}
}

func WithManifest(m *manifest) Option {
	return func(c *Connector) {
		c.manifest = m
	}
}

// id() returns a unique id for this connector based on the manifest.
// TODO: need to change id if the connector is used multiple times with different arguments
func (c *Connector) id() string {
	return fmt.Sprintf("%s-%s-%s-%s", c.manifest.Author, c.manifest.Name, c.manifest.Version, c.env)
}

// NewConsumerConnector creates a message queue consumer.
// Common overrideOpts are
//		kafka.ConfigMap{
//			"auto.offset.reset": "latest",
//		}
// for sink connectors to ignore all existing messages in the queue.

// Subscribe subscribes to a list of topics.
// To read:
//
//	sub, err := connector.Subscribe(...)
//	for msg := range sub {
//		// do something with the msg
//		// print(msg)
//
//		// commit to kafka to acknowledge receipt
//		consumer.CommitMessage(msg.Message)
//	}
func (c *Connector) Subscribe(topics []kafkautils.Topic, overrideOpts ...kafka.ConfigMap) (<-chan kafkautils.Message, error) {
	if !c.consumerStarted {
		err := c.startConsumer(overrideOpts...)
		if err != nil {
			return nil, err
		}
	}

	if err := c.SubscribeTopics(topics); err != nil {
		log.Error().Err(err).Msg("kafka subscribe proto error")
		return nil, err
	}

	return c.Consumer.Messages(), nil
}

func (c *Connector) SubscribeExample() error {
	env := "staging"

	sub, err := c.Subscribe(
		kafkautils.MustParseTopics([]string{
			".fct.nakji.common.0_0_0.market_trade",
			".fct.nakji.common.0_0_0.market_openinterest",
		}, env),
	)
	if err != nil {
		return err
	}

	// Listen to the sub channel for new messages
	go func() {
		for msg := range sub {
			// do something with the msg
			kafkautils.DebugPrint(msg.Topic, msg.Key, msg.ProtoMsg)

			// Commit to kafka to acknowledge receipt. Unnecessary if  `auto.offset.reset = latest` because you will want latest messages instead of from last commit.
			c.CommitMessage(msg.Message)
		}
	}()

	return nil
}

// ProduceMessage sends protobuf to message queue with a Topic and Key.
func (c *Connector) ProduceMessage(namespace, subject string, msgType kafkautils.MsgType, msg proto.Message) error {
	topic := c.generateTopicFromProto(msgType, msg)
	key := kafkautils.NewKey(namespace, subject)
	return c.ProduceMsg(context.TODO(), topic.String(), msg, key.Bytes(), nil)
}

// ProduceAndCommitMessage sends protobuf to message queue with a Topic and Key.
func (c *Connector) ProduceAndCommitMessage(namespace, subject string, msgType kafkautils.MsgType, msg proto.Message) error {
	if !c.producerStarted {
		err := c.startProducer()
		if err != nil {
			return err
		}
	}

	// Start a new transaction
	err := c.Producer.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	err = c.ProduceMessage(namespace, subject, msgType, msg)
	if err != nil {
		return err
	}

	err = c.Producer.CommitTransaction(context.TODO())
	if err != nil {
		log.Error().Err(err).Msg("Processor: Failed to commit transaction")

		err = c.Producer.AbortTransaction(context.TODO())
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}

	return nil
}

// startProducer() creates a new kafka producer with transactions enabled.
func (c *Connector) startProducer() error {
	txID := c.id()

	log.Info().
		Str("transactionID", txID).
		Msg("Initializing kafka producer")

	var err error
	c.Producer, err = kafkautils.NewProducer(c.kafkaUrl, txID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new kafka producer")
		return err
	}
	err = c.Producer.EnableTransactions()
	if err != nil {
		return err
	}
	c.producerStarted = true

	return nil
}

// startConsumer() creates a new kafka consumer and subscribes to a list of topics
func (c *Connector) startConsumer(overrideOpts ...kafka.ConfigMap) error {
	groupID := c.id()

	log.Info().
		Str("groupID", groupID).
		Msg("Initializing kafka consumer")

	// set auto.offset.reset for all topics because we don't care about the past for streams
	var err error
	c.Consumer, err = kafkautils.NewConsumer(c.kafkaUrl, groupID, overrideOpts...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create kafka consumer")
		return err
	}
	c.consumerStarted = true

	return nil
}

// generateTopicFromProto generates message queue topic names based on the protobuf message.
// Event names should be prefixed with contract_ or category_ when appropriate.
func (c *Connector) generateTopicFromProto(msgType kafkautils.MsgType, msg proto.Message) kafkautils.Topic {
	return kafkautils.NewTopic(
		c.env,
		msgType,
		c.manifest.Author,
		c.manifest.Name,
		c.manifest.Version.Version,
		msg,
	)
}

// RegisterProtos generates kafka topic and protobuf type mappings from proto.Message and registers them dynamically.
func (c *Connector) RegisterProtos(msgType kafkautils.MsgType, protos ...proto.Message) {
	if c.env == kafkautils.EnvDev {
		log.Debug().Msg("protoregistry is disabled in dev mode, set kafka.env to other values (e.g., test, staging) to enable it")
		return
	}

	tt := c.buildTopicTypes(msgType, protos...)

	err := protoregistry.RegisterDynamicTopics(c.protoRegistryHost, tt, msgType)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register dynamic topics")
	}
}

func (c *Connector) buildTopicTypes(msgType kafkautils.MsgType, protos ...proto.Message) protoregistry.TopicTypes {
	tt := make(map[string]proto.Message)

	for _, proto := range protos {
		tt[c.generateTopicFromProto(msgType, proto).Schema()] = proto
	}

	return tt
}

//	DEPRECATED, please use ProduceWithTransaction instead
//	initProduceChannel uses the incoming messages from protobuf message channel and forwards them to Kafka.
//	It wraps each message in a Kafka Transaction to ensure Exactly Once Semantics.
//	NOTE: this wraps individual messages with transactions so it adds a lot of overhead to kafka and reduces the usefulness of transactions
func (c *Connector) initProduceChannel(input <-chan *kafkautils.Message) {

	c.startProducer()
	ticker := time.NewTicker(time.Second)

	var messages []*kafkautils.Message

	for {
		select {
		case <-ticker.C:
			if len(messages) > 0 {
				c.ProduceWithTransaction(messages)
				messages = make([]*kafkautils.Message, 0)
			}
		case msg := <-input:
			messages = append(messages, msg)
		}
	}
}

//	ProduceWithTransaction wraps a slice of messages in a kafka transaction.
//	Produced messages will be pushed to kafka altogether or fail all at once.
func (c *Connector) ProduceWithTransaction(messages []*kafkautils.Message) error {
	if !c.producerStarted {
		c.startProducer()
	}

	err := c.Producer.BeginTransaction()
	if err != nil {
		log.Error().Err(err).
			Str("error code", err.(kafka.Error).Code().String()).
			Msg("failed to begin transaction")

		return err
	}

	for _, msg := range messages {

		topic := c.generateTopicFromProto(msg.MsgType, msg.ProtoMsg).String()
		// Get trace data from message
		ctx := trace.ContextWithSpan(context.TODO(), msg.Span)
		ctx = baggage.ContextWithBaggage(ctx, msg.Baggage)

		c.Producer.ProduceMsg(ctx, topic, msg.ProtoMsg, nil, nil)
	}

retry:
	err = c.Producer.CommitTransaction(context.TODO())
	if err != nil {
		if err.(kafka.Error).IsRetriable() {
			log.Warn().Err(err).
				Str("error code", err.(kafka.Error).Code().String()).
				Msg("failed to commit transactions, retrying..")
			goto retry

		} else if err.(kafka.Error).Code() == kafka.ErrProducerFenced {
			c.Producer.Close()

		} else if err.(kafka.Error).Code() == kafka.ErrTimedOut {
			log.Warn().Err(err).
				Str("error code", err.(kafka.Error).Code().String()).
				Msg("failed to commit transactions, timed out")

		} else {
			log.Error().Err(err).
				Str("error code", err.(kafka.Error).Code().String()).
				Msg("failed to commit transactions, aborting..")

			err = c.Producer.AbortTransaction(context.TODO())
			if err != nil {
				log.Fatal().Err(err).
					Str("error code", err.(kafka.Error).Code().String()).
					Msg("failed to abort transaction, killing producer..")
			}
		}
		return err
	}

	log.Debug().Msg("successfully committed transactions")

	return nil
}
