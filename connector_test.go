package connector

import (
	"testing"

	"github.com/Masterminds/semver"
	"github.com/nakji-network/connector/kafkautils"
)

func TestGenerateTopicFromProto(t *testing.T) {
	v, _ := semver.NewVersion("0.0.0")
	c := Connector{
		env: "test",
		manifest: &Manifest{
			Name:    "ethereum",
			Author:  "nakji",
			Version: version{Version: v},
		},
	}
	got := c.generateTopicFromProto(kafkautils.MsgTypeFct, &Transaction{}).String()
	want := "test.fct.nakji.ethereum.0_0_0.ethereum_Transaction"
	if got != want {
		t.Errorf("Error generating topic from proto: got=%q want=%q", got, want)
	}
}
