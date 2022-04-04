package connector

import (
	"testing"

	"github.com/Masterminds/semver"
)

func TestGenerateTopicFromProto(t *testing.T) {
	v, _ := semver.NewVersion("0.0.0")
	c := Connector{
		env: "test",
		manifest: &manifest{
			Name:    "ethereum",
			Author:  "nakji",
			Version: version{Version: v},
		},
	}
	got := c.GenerateTopicFromProto(&Transaction{}).String()
	want := "test.fct.nakji.ethereum.0_0_0.ethereum_Transaction"
	if got != want {
		t.Errorf("Error generating topic from proto: got=%q want=%q", got, want)
	}
}
