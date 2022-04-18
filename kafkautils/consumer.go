// https://github.com/confluentinc/confluent-kafka-go/blob/master/examples/idempotent_producer_example/idempotent_producer_example.go

package kafkautils

import (
	"context"
	"fmt"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	*kafka.Consumer
	MessageCh chan Message // kafka messages + Topic and proto message
}

func NewConsumer(brokers string, groupID string, overrideOpts ...kafka.ConfigMap) (*Consumer, error) {
	defaultOpts := kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		//"group.instance.id": groupInstanceID,
		"session.timeout.ms":              6000,
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
	c := Consumer{cons, make(chan Message, 1)}
	return &c, nil
}

func (c *Consumer) Messages() chan Message {
	return c.MessageCh
}

// Process kafka events and forward to Consumer.MessageCh
func (c *Consumer) SubscribeProto(ctx context.Context, topics []Topic) error {
	log.Info().Strs("topics", TopicsStrings(topics)).Msg("kafka subscribe")
	if err := c.SubscribeTopics(TopicsStrings(topics), nil); err != nil {
		log.Error().Err(err).Msg("kafka subscribe failure")
		return err
	}

	for {
		select {
		case ev := <-c.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				c.Assign(e.Partitions)
			case kafka.RevokedPartitions:
				fmt.Fprintf(os.Stderr, "%% %v\n", e)
				c.Unassign()
			case *kafka.Message:
				//log.Debug().Msgf("%% Message on %s:\n%s\n",
				//e.TopicPartition, string(e.Value))

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

				// If it exists in TopicTypeRegistry, use that; otherwise use Proto Registry
				if _, ok := TopicTypeRegistry[t.Schema()]; ok {
					protoMsg, err := t.UnmarshalProto(e.Value)
					if err != nil {
						log.Error().Err(err).Interface("topic", t).Msg("Unable to UnmarshalProto topic")
						continue
					}
					c.MessageCh <- Message{
						Message:  e,
						Topic:    t,
						Key:      k,
						ProtoMsg: protoMsg,
					}
				} else {
					dynamicProtoMsg, err := t.UnmarshalDynamicProto(e.Value)
					if err != nil {
						log.Error().Err(err).Interface("topic", t).Msg("Unable to UnmarshalProto topic")
						continue
					}
					c.MessageCh <- Message{
						Message:    e,
						Topic:      t,
						Key:        k,
						IsDynamic:  true,
						DynamicMsg: dynamicProtoMsg,
					}
				}
			case kafka.PartitionEOF:
				log.Info().Interface("event", e).Msg("%% Reached")
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}

		case <-ctx.Done():
			return nil
		}
	}
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
