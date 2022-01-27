package connector

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"
)

type result struct {
	Test version
}

func TestYamlUnmarshal(t *testing.T) {
	sVer := "1.1.1"
	r := &result{}
	err := yaml.Unmarshal([]byte(fmt.Sprintf("test: %q", sVer)), r)
	if err != nil {
		t.Errorf("Error unmarshaling version: %s", err)
	}
	got := r.Test.String()
	want := sVer
	if got != want {
		t.Errorf("Error unmarshaling unexpected object content: got=%q want=%q", got, want)
	}
}
