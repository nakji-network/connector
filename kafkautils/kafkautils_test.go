package kafkautils

import (
	"os"
	"testing"
)

// var mockTTR TTR

func TestMain(m *testing.M) {
	TopicTypeRegistry.Load(testTopicTypes)
	os.Exit(m.Run())
}
