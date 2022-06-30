package kafkautils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nakji-network/connector/monitor"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

//go:generate mockgen -destination=mocks/mock_producer.go -package=mocks . ProducerInterface
type ProducerInterface interface {
	InitTransactions(context.Context) error
	BeginTransaction() error
	WriteKafkaMessages(Topic, []byte, proto.Message) error
	CommitTransaction(context.Context) error
	AbortTransaction(context.Context) error
	EnableTransactions() error
	ProduceMsg(Topic, proto.Message, chan kafka.Event) error
	WriteAndCommitSink(<-chan *Message)
	WriteAndCommit(Topic, []byte, proto.Message) error
	MakeQueueTransactionSink() chan *Message
	Close()
}

type Producer struct {
	*kafka.Producer
	termChan chan bool
	doneChan chan bool
	closed   bool
}

//	Producer config flags and their default values can be found here
//	https://docs.confluent.io/platform/current/installation/configuration/producer-configs.html
const (
	//	the producer will wait for up to the given delay to allow other records to be sent so that the sends can be batched together
	KafkaProducerLingerMS = 1000

	// the maximum amount of time the client will wait for the response of a request
	KafkaProducerRequestTimeoutMS = 60000

	// (10mins) default 60000 (1min) https://docs.confluent.io/platform/current/installation/configuration/producer-configs.html
	KafkaProducerTransactionTimeoutMS = 600000

	//default 1000 https://docs.confluent.io/2.0.0/clients/librdkafka/CONFIGURATION_8md.html
	KafkaProducerQueueBufferingMaxMS = 2000
)

func MustNewProducer(brokers, transactionalID string) *Producer {
	p, err := NewProducer(brokers, transactionalID)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create kafka producer")
	}
	return p
}

// NewProducer produces new Kafka producer. Must call `EnableTransactions` before sending messages to start transactions.
func NewProducer(brokers, transactionalID string) (*Producer, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		//"enable.idempotence":     true,
		"linger.ms":              KafkaProducerLingerMS,
		"request.timeout.ms":     KafkaProducerRequestTimeoutMS,
		"transactional.id":       transactionalID,
		"transaction.timeout.ms": KafkaProducerTransactionTimeoutMS,
		"queue.buffering.max.ms": KafkaProducerQueueBufferingMaxMS,
		"compression.codec":      "snappy",
	})
	if err != nil {
		return nil, err
	}

	p := Producer{
		Producer: producer,
		// For signalling termination from main to go-routine
		termChan: make(chan bool, 1),
		// For signalling that termination is done from go-routine to main
		doneChan: make(chan bool),
	}

	// Go routine for serving the events channel for delivery reports and error events.
	go func() {
		defer close(p.doneChan)
		doTerm := false
		for !doTerm {
			select {
			case e := <-p.Events():
				switch ev := e.(type) {
				case *kafka.Message:
					// Message delivery report
					m := ev
					if m.TopicPartition.Error != nil {
						log.Error().
							Err(m.TopicPartition.Error).
							Msg("Delivery failed")
					} else {
						log.Debug().
							Str("offset", m.TopicPartition.Offset.String()).
							Str("topic", *m.TopicPartition.Topic).
							Bytes("key", m.Key).
							Int32("partition", m.TopicPartition.Partition).
							Msg("Delivered message")
					}

				case kafka.Error:
					// Generic client instance-level errors, such as
					// broker connection failures, authentication issues, etc.
					//
					// These errors should generally be considered informational
					// as the underlying client will automatically try to
					// recover from any errors encountered, the application
					// does not need to take action on them.
					//
					// But with idempotence enabled, truly fatal errors can
					// be raised when the idempotence guarantees can't be
					// satisfied, these errors are identified by
					// `e.IsFatal()`.

					e := ev
					if e.IsFatal() {
						// Fatal error handling.
						//
						// When a fatal error is detected by the producer
						// instance, it will emit kafka.Error event (with
						// IsFatal()) set on the Events channel.
						//
						// Note:
						//   After a fatal error has been raised, any
						//   subsequent Produce*() calls will fail with
						//   the original error code.
						log.Error().Err(e).Msg("FATAL ERROR: terminating")
						p.closed = true
						go p.Close()
					} else {
						log.Error().Err(e).Msg("")
					}

				default:
					log.Warn().Interface("event", ev).Msg("Ignored event")
				}

			case <-p.termChan:
				doTerm = true
			}
		}

	}()

	return &p, nil
}

// EnableTransactions enables transactions for this Kafka producer. Use this after everything is loaded but before sending any kafka messages
func (p *Producer) EnableTransactions() error {
	// Init Transactions within 2 minutes
	maxDuration := 120 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	err := p.InitTransactions(ctx)
	if err != nil {
		return err
	}

	return nil
}

// NOTE: must be last to be called in closing
func (p *Producer) Close() {
	// Clean termination to get delivery results
	// for all outstanding/in-transit/queued messages.
	log.Info().Msg("Flushing outstanding Kafka messages")
	p.Flush(15 * 1000)

	p.termChan <- true

	// wait for go-routine to terminate
	<-p.doneChan
	fatalErr := p.GetFatalError()

	// destroy transactional producer
	maxDuration := 10 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), maxDuration)
	defer cancel()

	err := p.Producer.AbortTransaction(ctx)
	if err != nil {
		if err.(kafka.Error).Code() == kafka.ErrState {
			// No transaction in progress, ignore the error.
			err = nil
		} else {
			log.Error().Err(err).Str("producer", p.String()).Msg("Failed to abort transaction")
		}
	}

	p.Producer.Close()

	// Exit application with an error (1) if there was a fatal error.
	if fatalErr != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

// https://github.com/confluentinc/confluent-kafka-go/blob/24f06a50dd915cc346c8a36e5a7f7306f4339cfe/examples/transactions_example/txnhelpers.go
// Commit offset position to a consumer when this producer transaction is successful.
func (p *Producer) SendOffsetsToTransaction(position kafka.TopicPartitions, c *Consumer) {
	consumerMetadata, err := c.GetConsumerGroupMetadata()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get consumer group metadata")
	}
	log.Info().
		Interface("pos", position).
		Interface("consumermetadata", consumerMetadata).
		Msg("attempting to send offsets")
	err = p.Producer.SendOffsetsToTransaction(nil, position, consumerMetadata)
	if err != nil {
		log.Error().Err(err).Msg("Processor: Failed to send offsets to transaction: aborting transaction")

		err = p.AbortTransaction(nil)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}

		// Rewind this input partition to the last committed offset.
		c.rewindPosition(position)
	} else {
		err = p.CommitTransaction(nil)
		if err != nil {
			log.Error().Err(err).Msg("Processor: Failed to commit transaction")

			err = p.AbortTransaction(nil)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}

			// Rewind this input partition to the last committed offset.
			c.rewindPosition(position)
		}
	}

	// Start a new transaction
	err = p.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

// WriteKafkaMessages writes plain kafka messages
func (p *Producer) WriteKafkaMessages(topic Topic, key []byte, value proto.Message) error {
	if p.closed {
		return fmt.Errorf("cannot write kafka message. Producer is already closed")
	}

	pbData, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	monitor.SetMetricsForKafkaLastWriteTime()

	topicString := topic.String()

	return p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicString, Partition: kafka.PartitionAny},
		Value:          pbData,
		Key:            key,
		//Timestamp:      time.Time,
	}, nil)
}

// WriteAndCommit writes for transactional producers, committing transaction after each write
func (p *Producer) WriteAndCommit(topic Topic, key []byte, value proto.Message) error {
	// Start a new transaction
	err := p.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	err = p.WriteKafkaMessages(topic, key, value)
	if err != nil {
		return err
	}

retry:
	err = p.CommitTransaction(nil)
	if err != nil {
		log.Error().Err(err).Msg("Processor: Failed to commit transaction")

		if err.(kafka.Error).IsRetriable() {
			time.Sleep(time.Second * 5)
			goto retry
		} else if err.(kafka.Error).TxnRequiresAbort() {
			err = p.AbortTransaction(nil)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to abort kafka transaction")
			}
		} else {
			log.Fatal().Err(err).Msg("failed to commit kafka transaction")
		}
	}

	return nil
}

func (p *Producer) WriteAndCommitSink(in <-chan *Message) {
	for msg := range in {
		err := p.WriteAndCommit(
			msg.Topic,
			msg.Key.Bytes(),
			msg.ProtoMsg,
		)
		if err != nil {
			log.Error().Err(err).
				Str("topic", msg.Topic.String()).
				Str("key", msg.Key.String()).
				Interface("protoMsg", msg.ProtoMsg).
				Msg("Write to kafka error")
		}
	}
}

// MakeQueueTransactionSink creates a channel that receives Kafka Messages. All messages within the channel are then automatically
// published to the specific topic in the `*kafkautils.Message`.
func (p *Producer) MakeQueueTransactionSink() chan *Message {
	sink := make(chan *Message, 10000)
	err := p.EnableTransactions()
	if err != nil {
		log.Fatal().Err(err).Msg("Transaction was not enabled")
	}
	go p.WriteAndCommitSink(sink)

	return sink
}

// ProduceMsg sends single message to Kafka
func (p *Producer) ProduceMsg(topic Topic, msg proto.Message, delivery chan kafka.Event) error {
	if p.closed {
		return fmt.Errorf("cannot produce message, producer is already closed")
	}

	pbData, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	monitor.SetMetricsForKafkaLastWriteTime()

	topicString := topic.String()

	return p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topicString, Partition: kafka.PartitionAny},
		Value:          pbData,
	}, delivery)
}
