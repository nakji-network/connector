package connector

import (
	"testing"

	"github.com/Masterminds/semver"
)

func TestGenerateTopicFromProto(t *testing.T) {
	v, _ := semver.NewVersion("0.0.0")
	c := Connector{
		manifest: &manifest{
			Name:    "ethereum",
			Author:  "nakji",
			Version: version{Version: v},
		},
	}
	got := c.GenerateTopicFromProto(&Transaction{})
	want := "nakji.ethereum.0_0_0.ethereum_transaction"
	if got != want {
		t.Errorf("Error generating topic from proto: got=%q want=%q", got, want)
	}
}
