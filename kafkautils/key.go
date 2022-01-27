// Helpers for Blep's kafka key naming scheme
// Design document: https://docs.google.com/spreadsheets/d/1PmYvbw8LiBYYooAINrm4_lGWiewKA-yq-zCGbXQVNfE/edit#gid=0
package kafkautils

import (
	"bytes"
	"fmt"
)

var KeyDelimiter = []byte(".")

// Kafka key struct. Keep lowercase.
type Key struct {
	Namespace []byte
	Subject   []byte
}

func NewKey(ns, s string) Key {
	return Key{
		[]byte(ns),
		[]byte(s),
	}
}

func (k Key) Bytes() []byte {
	return bytes.Join([][]byte{k.Namespace, k.Subject}, KeyDelimiter)
}

func (k Key) String() string {
	return string(k.Bytes())
}

func ParseKey(s []byte) (Key, error) {
	k := bytes.Split(s, KeyDelimiter)
	if len(k) != 2 {
		return Key{}, fmt.Errorf("Cannot parse key, needs 2 parts separated by \"%s\": %s", KeyDelimiter, string(s))
	}
	return Key{k[0], k[1]}, nil
}

// Return Key as slide of key/val tuples.
// eg. [[namespace, x], [subject, y]]
func (k Key) Tuple() [][]string {
	return [][]string{
		{"ns", string(k.Namespace)},
		{"s", string(k.Subject)},
	}
}
