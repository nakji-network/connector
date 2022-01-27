package connector

import (
	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v3"
)

type version struct {
	*semver.Version
}

func (v *version) UnmarshalYAML(b *yaml.Node) error {
	var s string
	if err := b.Decode(&s); err != nil {
		return err
	}
	temp, err := semver.NewVersion(s)
	if err != nil {
		return err
	}
	v.Version = temp
	return nil
}
