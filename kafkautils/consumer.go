// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/idempotent_producer_example/idempotent_producer_example.go

package kafkautils

import (
	"context"

	"github.com/nakji-network/connector/monitor"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type Consumer struct {
	*kafka.Consumer
	messages <-chan Message
}

const (
	spanName          = "kafka -> consumer"
	consumerEventName = "consume kafka message"
)

// NewConsumer prepares a message queue consumer. Subscribe to proto messages on .Messages chan
func NewConsumer(brokers string, groupID string, overrideOpts ...kafka.ConfigMap) (*Consumer, error) {
	defaultOpts := kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		//"group.instance.id": groupInstanceID,
		"session.timeout.ms":              30000, // default is at 45000
		"heartbeat.interval.ms":           2000,  // default is at 3000. must be set lower than session.timeout.ms, but typically should be set no higher than 1/3 of that value
		"enable.auto.commit":              false,
		"fetch.wait.max.ms":               500, // this is the default, maybe too slow
		"go.events.channel.enable":        true,
		"go.events.channel.size":          1,
		"go.application.rebalance.enable": true,
		"auto.offset.reset":               "earliest",       // maybe remove this for printer
		"isolation.level":                 "read_committed", //"read_committed"
	}

	// overwrite with optional opts
	if len(overrideOpts) > 0 {
		for k, v := range overrideOpts[0] {
			defaultOpts[k] = v
		}
	}

	log.Info().
		Interface("opts", defaultOpts).
		Msg("kafka connection opts")

	cons, err := kafka.NewConsumer(&defaultOpts)
	if err != nil {
		return nil, err
	}

	c := Consumer{Consumer: cons}
	c.messages = c.kafkaEventToProtoPipe(c.Events())

	return &c, nil
}

// Messages returns receiver channel for Kafka messages.
func (c *Consumer) Messages() <-chan Message {
	return c.messages
}

// SubscribeTopics subscribes to the provided list of topics. This replaces the current subscription.
func (c *Consumer) SubscribeTopics(topics []Topic) error {
	log.Debug().Strs("topics", TopicsStrings(topics)).Msg("kafka subscribe")
	return c.Consumer.SubscribeTopics(TopicsStrings(topics), nil)
}

// kafkaEventToProtoPipe converts and sends incoming kafka events to proto channel.
func (c *Consumer) kafkaEventToProtoPipe(in <-chan kafka.Event) <-chan Message {
	out := make(chan Message)
	go func() {
		for ev := range in {
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				log.Info().Interface("partitions", e.Partitions).Msg("kafka assigned partitions")
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				log.Info().Interface("partitions", e.Partitions).Msg("kafka revoked partitions")
				c.Unassign()
			case *kafka.Message:
				//log.Debug().Str("topicPartition", e.TopicPartition.String()).Str("val", string(e.Value)).Msg("kafka received message")

				// Extract tracing info from message
				ctx := otel.GetTextMapPropagator().Extract(context.Background(), monitor.NewMessageCarrier(e))

				tr := monitor.CreateTracer(monitor.DefaultTracerName)
				_, span := monitor.StartSpan(
					ctx,
					tr,
					spanName,
					trace.WithSpanKind(trace.SpanKindConsumer),
					trace.WithAttributes(
						semconv.MessagingSystemKey.String("kafka"),
						semconv.MessagingDestinationKindTopic,
						semconv.MessagingOperationProcess,
						semconv.MessagingKafkaMessageKeyKey.String(string(e.Key)),
						semconv.MessagingKafkaClientIDKey.String(c.String()),
						semconv.MessagingKafkaPartitionKey.Int(int(e.TopicPartition.Partition)),
						semconv.MessagingDestinationKey.String(*e.TopicPartition.Topic),
					),
				)
				span.AddEvent(consumerEventName)
				ctx = trace.ContextWithSpan(ctx, span)

				k, err := ParseKey(e.Key)
				if err != nil {
					log.Error().Err(err).Msg("")
					continue
				}

				t, err := ParseTopic(*e.TopicPartition.Topic)
				if err != nil {
					log.Error().Err(err).Str("topic", *e.TopicPartition.Topic).Msg("Unable to parse topic")
					continue
				}

				protoMsg, err := t.UnmarshalProto(e.Value)
				if err != nil {
					log.Error().Err(err).Interface("topic", t).Msg("Unable to UnmarshalProto topic")
					continue
				}
				out <- Message{
					Message:  e,
					Topic:    t,
					Key:      k,
					ProtoMsg: protoMsg,
					Context:  ctx,
				}

			case kafka.PartitionEOF:
				log.Info().Str("topic", *e.Topic).Msg("EOF Reached")
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				log.Warn().Err(e).Msg("")
			}
		}
		close(out)
	}()
	return out
}

// rewindConsumerPosition rewinds the consumer to the last committed offset or
// the beginning of the partition if there is no committed offset.
// This is to be used when the current transaction is aborted.
func (c *Consumer) rewindPosition(position kafka.TopicPartitions) {
	committed, err := c.Committed(position, 10*1000 /* 10s */)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	for _, tp := range committed {
		if tp.Offset < 0 {
			// No committed offset, reset to earliest
			tp.Offset = kafka.OffsetBeginning
		}

		log.Info().Int32("partition", tp.Partition).Interface("offset", tp.Offset).Msg("Processor: rewinding input partition to offset")

		err = c.Seek(tp, -1)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}
}
