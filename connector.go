package connector

import (
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/nakji-network/connector/config"
	"github.com/nakji-network/connector/kafkautils"
	"github.com/nakji-network/connector/monitor"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

type Connector struct {
	manifest *manifest
	Config *viper.Viper // TODO: maybe our own config struct for globals and nonglobals. Namespace so it can't read configs from other connectors (secrets)

	*kafkautils.Producer
	*kafkautils.Consumer

	producerStarted bool
	consumerStarted bool
}

func NewConnector() *Connector {
	c := &Connector{
		manifest: LoadManifest(),
		Config: config.GetConfig(),
	}

	log.Info().
		Str("id", c.id()).
		Msg("Starting connector")

	monitor.StartMonitor(c.id())

	return c
}

// id() returns a unique id for this connector based on the manifest.
// TODO: need to change id if the connector is used multiple times with different arguments
func (c *Connector) id () string {
	return fmt.Sprintf("%s-%s-%s", c.manifest.Author, c.manifest.Name, c.manifest.Version)
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
func (c *Connector) Subscribe(topics []kafkautils.Topic, overrideOpts ...kafka.ConfigMap) (chan kafkautils.Message, error) {
	if !c.consumerStarted {
		err := c.startConsumer(overrideOpts...)
		if err != nil {
			return nil, err
		}
		c.consumerStarted = true
	}

	err := c.SubscribeProto(topics)
	if err != nil {
		log.Error().Err(err).Msgf("kafka subscribe proto error")
		return nil, err
	}

	return c.Consumer.MessageCh, nil
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

// startProducer() creates a new kafka producer with transactions enabled.
func (c *Connector) startProducer() error {
	txID := c.id()

	log.Info().
		Str("transactionID", txID).
		Msg("Initializing kafka producer")

	var err error
	c.Producer, err = kafkautils.NewProducer(c.Config.GetString("kafka.url"), txID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create new kafka producer")
		return err
	}
	c.Producer.EnableTransactions()

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
	c.Consumer, err = kafkautils.NewConsumer(c.Config.GetString("kafka.url"), groupID, overrideOpts...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create kafka consumer")
		return err
	}

	return nil
}

// GenerateTopicFromProto generates message queue topic names based on the protobuf message.
// Event names should be prefixed with contract_ or category_ when appropriate.
func (c *Connector) GenerateTopicFromProto(msg proto.Message) string {
	author := c.manifest.Author
	connectorName := c.manifest.Name
	version := strings.ReplaceAll(c.manifest.Version.String(), ".", "_")
	eventName := strings.ToLower(string(msg.ProtoReflect().Descriptor().Name()))

	return strings.Join([]string{author, connectorName, version, eventName}, ".")
}
