package kafkautils

import (
	"os"
	"testing"
)

var mockTTR TTR

func TestMain(m *testing.M) {
	mockTTR = make(TTR)
	mockTTR.Load(testTopicTypes)
	os.Exit(m.Run())
}
