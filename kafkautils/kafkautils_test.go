package kafkautils

import "testing"

var mockTTR TTR

func TestMain(*testing.M) {
	mockTTR = make(TTR)
	mockTTR.Load(testTopicTypes)
}
