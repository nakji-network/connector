package connector

import (
	"fmt"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/heptiolabs/healthcheck"
	"github.com/nakji-network/connector/protoregistry"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"

	"github.com/nakji-network/connector/chain"
	"github.com/nakji-network/connector/config"
	"github.com/nakji-network/connector/kafkautils"
	"github.com/nakji-network/connector/monitor"
)

type Connector struct {
	manifest *manifest
	Config   *viper.Viper
	Health   healthcheck.Handler

	env             kafkautils.Env
	MsgType         kafkautils.MsgType
	kafkaUrl        string
	producerStarted bool
	consumerStarted bool
	*kafkautils.Producer
	*kafkautils.Consumer

	ChainClients     *chain.Clients
	ProtoRegistryCli *protoregistry.Client
}

// NewConnector returns a base connector implementation that other connectors can embed to add on to.
func NewConnector(path string) *Connector {
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
		manifest:         LoadManifest(path),
		env:              kafkautils.Env(conf.GetString("kafka.env")),
		MsgType:          kafkautils.Fct,
		kafkaUrl:         conf.GetString("kafka.url"),
		ChainClients:     chain.NewClients(rpcMap),
		Health:           healthcheck.NewHandler(),
		ProtoRegistryCli: prc,
	}

	c.Config = conf

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

// Subscribe creates a message queue consumer and subscribes to a list of topics.
// Common overrideOpts are
//		kafka.ConfigMap{
//			"auto.offset.reset": "latest",
//		}
// for sink connectors to ignore all existing messages in the queue.

// To read:
// 	sub, err := connector.Subscribe(...)
// 	for msg := range sub {
// 		// do something with the msg
// 		// print(msg)
//
// 		// commit to kafka to acknowledge receipt
// 		consumer.CommitMessage(msg.Message)
// 	}
func (c *Connector) Subscribe(topics []kafkautils.Topic, overrideOpts ...kafka.ConfigMap) (<-chan kafkautils.Message, error) {
	if !c.consumerStarted {
		err := c.startConsumer(overrideOpts...)
		if err != nil {
			return nil, err
		}
		c.consumerStarted = true
	}

	if err := c.SubscribeTopics(topics); err != nil {
		log.Error().Err(err).Msg("kafka subscribe proto error")
		return nil, err
	}

	return c.Consumer.Messages, nil
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

// MakeQueueTransactionSink creates a channel of kafka Messages. All messages sent to
// this channel will be wrapped in a transaction and published to the message queue.
func (c *Connector) MakeQueueTransactionSink() (chan *kafkautils.Message, error) {
	if !c.producerStarted {
		err := c.startProducer()
		if err != nil {
			return nil, err
		}
		c.producerStarted = true
	}

	sink := make(chan *kafkautils.Message, 10000)
	go c.Producer.WriteAndCommitSink(sink)

	return sink, nil
}

// ProduceMessage sends protobuf to message queue with a Topic and Key.
func (c *Connector) ProduceMessage(namespace, subject string, msg proto.Message) error {
	if !c.producerStarted {
		err := c.startProducer()
		if err != nil {
			return err
		}
		c.producerStarted = true
	}

	topic := c.GenerateTopicFromProto(msg)
	key := kafkautils.NewKey(namespace, subject)
	return c.WriteKafkaMessages(topic, key.Bytes(), msg)
}

// ProduceMessage sends protobuf to message queue with a Topic and Key.
func (c *Connector) ProduceAndCommitMessage(namespace, subject string, msg proto.Message) error {
	err := c.ProduceMessage(namespace, subject, msg)
	if err != nil {
		return err
	}

	err = c.Producer.CommitTransaction(nil)
	if err != nil {
		log.Error().Err(err).Msg("Processor: Failed to commit transaction")

		err = c.Producer.AbortTransaction(nil)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}
	// Start a new transaction
	err = c.Producer.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
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

// GenerateTopicFromProto generates message queue topic names based on the protobuf message.
// Event names should be prefixed with contract_ or category_ when appropriate.
func (c *Connector) GenerateTopicFromProto(msg proto.Message) kafkautils.Topic {
	return kafkautils.NewTopic(
		c.env,
		c.MsgType,
		c.manifest.Author,
		c.manifest.Name,
		c.manifest.Version.Version,
		msg)
}
