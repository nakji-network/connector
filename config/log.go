package config

import (
	"os"
	"time"

	fluentd "github.com/joonix/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Version string
var Build string
var BuildTime string

var logrusLevels = map[string]logrus.Level{
	"debug": logrus.DebugLevel,
}
var zerologLevels = map[string]zerolog.Level{
	"debug": zerolog.DebugLevel,
	"info":  zerolog.InfoLevel,
	"fatal": zerolog.FatalLevel,
}

// InitLogrus configures the global log.
func InitLogging(v *viper.Viper) {
	// Logrus (legacy)
	if level, ok := logrusLevels[v.GetString("log.level")]; ok {
		logrus.SetLevel(level)
	}

	switch v.GetString("log.format") {
	case "JSON":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "FLUENTD":
		logrus.SetFormatter(fluentd.NewFormatter())
	}

	// Zerolog
	if level, ok := zerologLevels[v.GetString("log.level")]; ok {
		zerolog.SetGlobalLevel(level)
		log.Logger = log.With().Caller().Logger()

		if v.GetString("log.format") == "console" {
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}
	}
	// Zerolog for gcp
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	log.Info().
		Str("Version", Version).
		Str("Build", Build).
		Str("BuildTime", BuildTime).
		Msg("")
}
