// This file handles the manifest.yaml file that describes a connector's metadata.
package connector

import (
	"os"

	"github.com/Masterminds/semver"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Name         string
	Author       string
	Version      version
	Url          string `yaml:"url,omitempty"`
	PrimaryColor string `yaml:"primaryColor,omitempty"`
	Links        link   `yaml:"links,omitempty"`
}

type link struct {
	Github   string `yaml:"github,omitempty"`
	Twitter  string `yaml:"twitter,omitempty"`
	Discord  string `yaml:"discord,omitempty"`
	Telegram string `yaml:"telegram,omitempty"`
	Medium   string `yaml:"medium,omitempty"`
}

type ManifestOption func(*Manifest)

func parseManifestOptions(m *Manifest, options ...ManifestOption) {
	for _, option := range options {
		option(m)
	}
}

func WithUrl(url string) ManifestOption {
	return func(m *Manifest) {
		m.Url = url
	}
}

func WithPrimaryColor(primaryColor string) ManifestOption {
	return func(m *Manifest) {
		m.PrimaryColor = primaryColor
	}
}

func WithLinks(links link) ManifestOption {
	return func(m *Manifest) {
		m.Links = links
	}
}

// TODO: tell user to use embed to embed the manifest.yaml file or else they'll have to manually keep the file with the exe
func LoadManifest() *Manifest {
	log.Info().Msg("Loading Manifest")

	yfile, err := os.ReadFile("Manifest.yaml")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to open file Manifest.yaml.")
		return nil
	}

	m := new(Manifest)

	err2 := yaml.Unmarshal(yfile, m)
	if err2 != nil {
		log.Fatal().Err(err2).Msg("Failed to read yaml from Manifest.yaml.")
	}

	if m.Name == "" || m.Author == "" || m.Version.Version == nil {
		log.Fatal().Msg("Missing name, author, and version fields from Manifest.yaml.")
	}

	log.Info().
		Str("name", m.Name).
		Str("author", m.Author).
		Str("version", m.Version.String()).
		Msg("Manifest loaded")

	return m
}

func NewManifest(name string, author string, ver string, options ...ManifestOption) *Manifest {
	nv, err := semver.NewVersion(ver)
	if err != nil {
		log.Fatal().Err(err).Msg("invalid version")
	}

	m := &Manifest{
		Name:    name,
		Author:  author,
		Version: version{nv},
	}

	parseManifestOptions(m, options...)

	return m
}
