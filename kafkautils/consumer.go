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
	c := Consumer{cons, make(chan Message, 1)}
	return &c, nil
}

func (c *Consumer) Messages() chan Message {
	return c.MessageCh
}

// Process kafka events and forward to Consumer.MessageCh
func (c *Consumer) SubscribeProto(topics []string) error {
	log.Info().Strs("topics", topics).Msg("kafka subscribe")
	err := c.SubscribeTopics(topics, nil)
	if err != nil {
		log.Error().Msg("kafka subscribe failure")
		return err
	}
	return nil
}

//	Listen will forward incoming events to message channel. It should be called on a separate goroutine.
func (c *Consumer) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("consumer listening to kafka is cancelled")
			return
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

			case kafka.PartitionEOF:
				log.Info().Interface("event", e).Msg("%% Reached")
			case kafka.Error:
				// Errors should generally be considered as informational, the client will try to automatically recover
				fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
			}
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
