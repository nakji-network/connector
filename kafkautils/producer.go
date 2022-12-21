package kafkautils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nakji-network/connector/monitor"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type Producer struct {
	*kafka.Producer
	termChan chan bool
	doneChan chan bool
	closed   bool
}

// Producer config flags and their default values can be found here
// https://docs.confluent.io/platform/current/installation/configuration/producer-configs.html
const (
	//	the producer will wait for up to the given delay to allow other records to be sent so that the sends can be batched together
	KafkaProducerLingerMS = 1000

	// the maximum amount of time the client will wait for the response of a request
	KafkaProducerRequestTimeoutMS = 60000

	// (10mins) default 60000 (1min) https://docs.confluent.io/platform/current/installation/configuration/producer-configs.html
	KafkaProducerTransactionTimeoutMS = 600000

	//default 1000 https://docs.confluent.io/2.0.0/clients/librdkafka/CONFIGURATION_8md.html
	KafkaProducerQueueBufferingMaxMS = 2000

	producerEventName = "produce kafka message"
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
		// "enable.idempotence":     true,
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

// DEPRECATED, this function will be removed in a future release. Please use ProduceWithTransaction in connector.go instead.
// WriteAndCommit writes for transactional producers, committing transaction after each write
func (p *Producer) WriteAndCommit(ctx context.Context, topic Topic, key []byte, value proto.Message) error {
	err := p.ProduceMsg(ctx, topic.String(), value, key, nil)
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

	err = p.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	return nil
}

const txBatchSize = 50

// DEPRECATED, this function will be removed in a future release. Please use ProduceWithTransaction in connector.go instead.
// WriteAndCommitSink will wrap messages in a transaction. The transaction is committed
// once there are 10 messages in the transaction or if the interval timer has been hit.
func (p *Producer) WriteAndCommitSink(in <-chan *Message) {
	ticker := time.NewTicker(time.Second)
	msgCounter := 0

	// Start a new transaction
	err := p.BeginTransaction()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	for msg := range in {
		ctx := trace.ContextWithSpan(context.TODO(), msg.Span)
		ctx = baggage.ContextWithBaggage(ctx, msg.Baggage)
		select {
		case <-ticker.C:
			if msgCounter > 0 {
				err := p.WriteAndCommit(
					ctx,
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

				msgCounter = 0
			}

		default:
			if msgCounter >= txBatchSize {
				err := p.WriteAndCommit(
					ctx,
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

				msgCounter = 0
				continue
			}

			err := p.ProduceMsg(ctx, msg.Topic.String(), msg.ProtoMsg, msg.Key.Bytes(), nil)
			if err != nil {
				log.Error().Err(err).
					Str("topic", msg.Topic.String()).
					Str("key", msg.Key.String()).
					Interface("protoMsg", msg.ProtoMsg).
					Msg("Write to kafka error")
			}
			msgCounter++
		}
	}
}

// DEPRECATED, this function will be removed in a future release. Please use ProduceWithTransaction in connector.go instead.
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

// ProduceMsg sends single protobuf message to a Kafka topic. It can optionally include a key and timestamp.
// An event channel can be provided for Kafka event response.
func (p *Producer) ProduceMsg(ctx context.Context, topic string, msg proto.Message, key []byte, delivery chan kafka.Event) error {
	if p.closed {
		return fmt.Errorf("cannot produce message, producer is already closed")
	}

	pbData, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	monitor.SetMetricsForKafkaLastWriteTime()

	kafkaMsg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          pbData,
	}

	if key != nil {
		kafkaMsg.Key = key
	}

	if ctx != nil {
		// Get span from context
		span := trace.SpanFromContext(ctx)

		// Add kafka produce time to baggage
		ctx, _ := monitor.NewLatencyBaggage(ctx, monitor.LatencyKafkaProduceKey, time.Now())

		// Inject trace metadata into kafka message headers
		otel.GetTextMapPropagator().Inject(ctx, monitor.NewMessageCarrier(kafkaMsg))

		span.AddEvent(producerEventName)
		monitor.EndSpan(span, nil)
	}

	return p.Produce(kafkaMsg, delivery)
}

// ListenDeliveryChan processes acknowledgements from Kafka broker upon Produce() calls.
// Using this method will disable global event processing by the producer.
func (p *Producer) ListenDeliveryChan(delivery chan kafka.Event) {
	for e := range delivery {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				err := ev.TopicPartition.Error
				log.Warn().Err(err).
					Str("error code", err.(kafka.Error).Code().String()).
					Interface("partition", ev.TopicPartition).
					Msg("failed to deliver message")

			} else {
				log.Debug().
					Str("topic", *ev.TopicPartition.Topic).
					Int32("partition", ev.TopicPartition.Partition).
					Interface("offset", ev.TopicPartition.Offset).
					Msg("successfully produced record")
			}
		}
	}
}
