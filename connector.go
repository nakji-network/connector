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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Connector struct {
	manifest *manifest
	Config   *viper.Viper
	Health   healthcheck.Handler

	env      kafkautils.Env
	MsgType  kafkautils.MsgType
	kafkaUrl string
	kafkautils.ProducerInterface
	*kafkautils.Consumer

	ChainClients     *chain.Clients
	ProtoRegistryCli *protoregistry.Client
}

func NewProducerConnector() (*Connector, error) {
	c := newConnector()
	err := c.startProducer()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// NewConsumerConnector creates a message queue consumer.
// Common overrideOpts are
//		kafka.ConfigMap{
//			"auto.offset.reset": "latest",
//		}
// for sink connectors to ignore all existing messages in the queue.
func NewConsumerConnector(overrideOpts ...kafka.ConfigMap) (*Connector, error) {
	c := newConnector()
	err := c.startConsumer(overrideOpts...)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// newConnector returns a base connector implementation that other connectors can embed to add on to.
func newConnector() *Connector {
	conf := config.GetConfig()
	conf.SetDefault("kafka.env", "dev")
	conf.SetDefault("protoregistry.host", "localhost:9191")

	rpcMap := make(map[string]chain.RPCs)
	err := conf.UnmarshalKey("rpcs", &rpcMap)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not load RPC list from config file")
	}

	// Create a proto registry client
	prc := protoregistry.NewClient(conf.GetString("protoregistry.host"))

	c := &Connector{
		manifest:         LoadManifest(),
		env:              kafkautils.Env(conf.GetString("kafka.env")),
		MsgType:          kafkautils.Fct,
		kafkaUrl:         conf.GetString("kafka.url"),
		ChainClients:     chain.NewClients(rpcMap),
		Health:           healthcheck.NewHandler(),
		ProtoRegistryCli: prc,
	}

	c.Config = conf.Sub(c.id())
	if c.Config == nil {
		c.Config = viper.New()
	}

	log.Info().
		Str("id", c.id()).
		Msg("Starting connector")

	// For Liveness and Readiness Probe checks
	go http.ListenAndServe("0.0.0.0:8080", c.Health)
	log.Info().Str("addr", "0.0.0.0:8080").Msg("healthcheck listening on /live and /ready")

	monitor.StartMonitor(c.id())

	return c
}

// id() returns a unique id for this connector based on the manifest.
// TODO: need to change id if the connector is used multiple times with different arguments
func (c *Connector) id() string {
	return fmt.Sprintf("%s-%s-%s-%s", c.manifest.Author, c.manifest.Name, c.manifest.Version, c.env)
}

// Subscribe subscribes to a list of topics.
// To read:
// 	sub, err := connector.Subscribe(...)
// 	for msg := range sub {
// 		// do something with the msg
// 		// print(msg)
//
// 		// commit to kafka to acknowledge receipt
// 		consumer.CommitMessage(msg.Message)
// 	}
func (c *Connector) Subscribe(topics []kafkautils.Topic) (<-chan kafkautils.Message, error) {
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
func (c *Connector) ProduceMessage(namespace, subject string, msg proto.Message) error {
	topic := c.generateTopicFromProto(msg)
	key := kafkautils.NewKey(namespace, subject)
	return c.WriteKafkaMessages(topic, key.Bytes(), msg)
}

// ProduceAndCommitMessage sends protobuf to message queue with a Topic and Key.
func (c *Connector) ProduceAndCommitMessage(namespace, subject string, msg proto.Message) error {
	// Start a new transaction
	err := c.ProducerInterface.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	err = c.ProduceMessage(namespace, subject, msg)
	if err != nil {
		return err
	}

	err = c.ProducerInterface.CommitTransaction(nil)
	if err != nil {
		log.Error().Err(err).Msg("Processor: Failed to commit transaction")

		err = c.ProducerInterface.AbortTransaction(nil)
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
	c.ProducerInterface, err = kafkautils.NewProducer(c.kafkaUrl, txID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new kafka producer")
		return err
	}
	err = c.ProducerInterface.EnableTransactions()
	if err != nil {
		return err
	}

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

	return nil
}

// generateTopicFromProto generates message queue topic names based on the protobuf message.
// Event names should be prefixed with contract_ or category_ when appropriate.
func (c *Connector) generateTopicFromProto(msg proto.Message) kafkautils.Topic {
	return kafkautils.NewTopic(
		c.env,
		c.MsgType,
		c.manifest.Author,
		c.manifest.Name,
		c.manifest.Version.Version,
		msg,
	)
}

// RegisterProtos generates kafka topic and protobuf type mappings from proto.Message and registers them dynamically.
func (c *Connector) RegisterProtos(protos ...proto.Message) {
	tt := c.buildTopicTypes(protos...)

	err := c.ProtoRegistryCli.RegisterDynamicTopics(tt, c.MsgType)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to register dynamic topics")
	}
}

func (c *Connector) buildTopicTypes(protos ...proto.Message) protoregistry.TopicTypes {
	tt := make(map[string]proto.Message)

	for _, proto := range protos {
		tt[c.generateTopicFromProto(proto).Schema()] = proto
	}

	return tt
}

func (c *Connector) InitProduceChannel(input <-chan protoreflect.ProtoMessage) {
	duration := time.Second * 1

	ticker := time.NewTicker(duration)
	delivery := make(chan kafka.Event)
	defer close(delivery)

	go c.ProducerInterface.ListenDeliveryChan(delivery)

start:
	hasMessage := false

	err := c.ProducerInterface.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).
			Str("error code", err.(kafka.Error).Code().String()).
			Msg("failed to begin transaction")
	}

	for msg := range input {
		select {
		case <-ticker.C:

		retry:
			if hasMessage {

				err := c.ProducerInterface.CommitTransaction(context.TODO())
				if err != nil {
					if err.(kafka.Error).IsRetriable() {
						log.Warn().Err(err).
							Str("error code", err.(kafka.Error).Code().String()).
							Msg("failed to commit transactions, retrying..")
						goto retry

					} else if err.(kafka.Error).Code() == kafka.ErrProducerFenced {
						c.ProducerInterface.Close()

					} else if err.(kafka.Error).Code() == kafka.ErrTimedOut {
						log.Warn().Err(err).
							Str("error code", err.(kafka.Error).Code().String()).
							Msg("failed to commit transactions, timed out")

					} else {
						log.Error().Err(err).
							Str("error code", err.(kafka.Error).Code().String()).
							Msg("failed to commit transactions, aborting..")

						err = c.ProducerInterface.AbortTransaction(context.TODO())
						if err != nil {
							log.Fatal().Err(err).
								Str("error code", err.(kafka.Error).Code().String()).
								Msg("failed to abort transaction, killing producer..")
						}
					}
				}

				ticker.Reset(duration)

				goto start
			}
		default:

			hasMessage = true
			topic := c.generateTopicFromProto(msg)
			c.ProducerInterface.ProduceMsg(topic, msg, nil, time.Time{}, delivery)
		}
	}
}
