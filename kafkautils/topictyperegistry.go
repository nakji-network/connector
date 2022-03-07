package kafkautils

import (
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

// TODO: convert this into a schemaregistry type microservice.
// But it's hard to load proto/struct in over http

type TTR map[string]proto.Message

var TopicTypeRegistry = TTR{
	// Examples:
	//"kline": &common.Kline{},
	//"block-*": &ethereum.Block{},
}

// Get schema from TTR. Schemas with an interval (eg block-1m) defined get
// matched to * interval (eg block-*) to maintain compatibility if interval
// requirements change.
func (r TTR) Get(schema string) proto.Message {
	s := strings.SplitN(schema, "-", 2)
	switch len(s) {
	case 1:
		return r[schema]
	case 2:
		return r[s[0]+"-*"]
	}
	return nil
}

// Set schema in TTR. Fatal error if invalid key.
func (r TTR) Set(key string, value proto.Message) {
	match, err := regexp.MatchString(`^[a-z]*(?:-\*)?$`, key)
	if err != nil || !match {
		log.Fatalf("Attempted to set invalid topic: %s", key)
	}
	TopicTypeRegistry[key] = value
}

// Load an external map[string]proto.Message
func (r TTR) Load(b map[string]proto.Message) {
	for k, v := range b {
		r[k] = v
	}
}

func GetActiveSchemas(queryParams []string) map[string]bool {
	schemas := make(map[string]bool)
	if len(queryParams) == 0 {
		for k := range TopicTypeRegistry {
			schemas[k] = true
		}

	} else {
		for _, query := range queryParams {
			stream, err := NewSchema(query)
			if err != nil {
				continue
			}

			for k := range TopicTypeRegistry {
				if stream.hasSchema(k) {
					schemas[k] = true
				}
			}
		}
	}

	return schemas
}
