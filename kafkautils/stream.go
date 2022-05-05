package kafkautils

import (
	"fmt"
	"strings"
)

type StreamName struct {
	Author    string //	e.g. nakji
	Namespace string //	e.g. common
	Version   string // e.g. 0_0_0
	Contract  string //	e.g. market
	Event     string //	e.g. trade
	Period    string // e.g. *
}

func NewSchema(schema string) (*StreamName, error) {
	segments := strings.Split(schema, TopicContextSeparator)
	if len(segments) != TopicNumSegments {
		return nil, fmt.Errorf(fmt.Sprintf("incorrect number of segments. want: %d, got: %d", TopicNumSegments, len(segments)))
	}

	var contractEvent []string
	if segments[3] == "*" {
		contractEvent = []string{"*", "*"}
	} else {
		contractEvent = strings.Split(segments[3], TopicContractSeparator)
		if len(contractEvent) != 2 {
			return nil, fmt.Errorf(fmt.Sprintf("incorrect schema definition. want: contract_event, got: %s", segments[3]))
		}
	}

	s := &StreamName{
		Author:    segments[0],
		Namespace: segments[1],
		Version:   segments[2],
		Contract:  contractEvent[0],
		Event:     contractEvent[1],
	}

	aggregation := strings.SplitAfterN(contractEvent[1], TopicAggregateSeparator, 2)
	if len(aggregation) > 1 {
		s.Period = aggregation[1]
	}

	if !s.isValid() {
		return nil, fmt.Errorf(fmt.Sprintf("%s is not a valid stream", schema))
	}

	return s, nil
}

func (s *StreamName) isValid() bool {
	//	TODO Expand input validation
	return s.Author != "" && s.Namespace != "" && s.Version != "" && s.Contract != "" && s.Event != ""
}

func (s *StreamName) hasSchema(schema string) bool {
	in, err := NewSchema(schema)
	if err != nil {
		return false
	}

	return (s.Author == "*" || s.Author == in.Author) &&
		(s.Namespace == "*" || s.Namespace == in.Namespace) &&
		(s.Version == "*" || s.Version == in.Version) &&
		(s.Contract == "*" || s.Contract == in.Contract) &&
		(s.Event == "*" || s.Event == in.Event)
}
