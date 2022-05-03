// This file handles the manifest.yaml file that describes a connector's metadata.
package connector

import (
	"io/ioutil"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type manifest struct {
	Name    string
	Author  string
	Version version
}

// TODO: tell user to use embed to embed the manifest.yaml file or else they'll have to manually keep the file with the exe

func LoadManifest() *manifest {
	log.Info().Msg("Loading Manifest")

	yfile, err := ioutil.ReadFile("manifest.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open file manifest.yaml.")
	}

	m := new(manifest)

	err2 := yaml.Unmarshal(yfile, m)
	if err2 != nil {
		log.Fatal().Err(err2).Msg("Failed to read yaml from manifest.yaml.")
	}

	if m.Name == "" || m.Author == "" || m.Version.Version == nil {
		log.Fatal().Msg("Missing name, author, and version fields from manifest.yaml.")
	}

	log.Info().
		Str("name", m.Name).
		Str("author", m.Author).
		Str("version", m.Version.String()).
		Msg("Manifest loaded")

	return m
}
