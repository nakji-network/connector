// NOTE: Deprecated, do not use for any new code
// Helpers for Blep's kafka key naming scheme
// Design document: https://docs.google.com/spreadsheets/d/1PmYvbw8LiBYYooAINrm4_lGWiewKA-yq-zCGbXQVNfE/edit#gid=0
package kafkautils

import (
	"fmt"
	"strings"
)

const KeyDelimiter = "."

// Kafka key struct. Keep lowercase.
type Key struct {
	Namespace string `json:"namespace"`
	Subject   string `json:"subject"`
}

func NewKey(ns, s string) Key {
	return Key{
		ns,
		s,
	}
}

func (k Key) Bytes() []byte {
	return []byte(k.String())
}

func (k Key) String() string {
	return strings.Join([]string{k.Namespace, k.Subject}, KeyDelimiter)
}

func ParseKey(s []byte) (Key, error) {
	k := strings.Split(string(s), KeyDelimiter)
	if len(k) != 2 {
		return Key{}, fmt.Errorf("Cannot parse key, needs 2 parts separated by \"%s\": %s", KeyDelimiter, string(s))
	}
	return Key{k[0], k[1]}, nil
}

// Return Key as slide of key/val tuples.
// eg. [[namespace, x], [subject, y]]
func (k Key) Tuple() [][]string {
	return [][]string{
		{"ns", k.Namespace},
		{"s", k.Subject},
	}
}
