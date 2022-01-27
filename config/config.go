// Package config Handles all config duties. Inspired by https://github.com/dunglas/mercure
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//	Wrapper for config for testing
type IConfig interface {
	SetDefault(string, interface{})
	GetString(string) string
	GetInt64(string) int64
}

var ErrInvalidConfig = errors.New("invalid config")
var conf *viper.Viper

func GetConfig() *viper.Viper {
	if conf == nil {
		conf = InitConfig()
	}

	return conf
}

// SetConfigDefaults sets defaults on a Viper instance.
func SetConfigDefaults(v *viper.Viper) {
	v.SetDefault("debug", false)
	//v.SetDefault("kafka.url", )
	//v.SetDefault("kafka.topic", "Topic1")
	v.SetDefault("ethereum.kafka.txID", "ethereum-source")
	v.SetDefault("streamserver.addr", ":8080")
	v.SetDefault("websocket.wcReplicationFactor", 1)
	v.SetDefault("websocket.hcReplicationFactor", 1)
	v.SetDefault("websocket.maxIdleSecond", 60)
	v.SetDefault("prometheus.metrics.port", 9999)
	v.SetDefault("websocket.coolDownSecond", 30)
	v.SetDefault("websocket.lruCacheSize", 128)
	v.SetDefault("websocket.allowedBlocksBehind", 3)
	v.SetDefault("prometheus.metrics.port", 9999)
}

// ValidateConfig validates a Viper instance.
func ValidateConfig(v *viper.Viper) error {
	if v.GetString("kafka.url") == "" {
		return fmt.Errorf(`%w: "kafka.url" configuration parameter must be defined`, ErrInvalidConfig)
	}
	if v.GetString("aws.id") == "" || v.GetString("aws.secret") == "" {
		fmt.Println("AWS credentials not present in config yaml")
	}
	return nil
}

// SetFlags creates flags and bind them to Viper.
func SetFlags(v *viper.Viper) {
	pflag.BoolP("debug", "d", false, "enable the debug mode")
	pflag.StringP("kafka.url", "k", "", "kafka url")
	pflag.String("log.format", "", "log format: \"\", JSON, FLUENTD")
	pflag.String("streamserver.addr", "", "streamserver listener address")

	// TODO: refactor it: should call parse() when actually need it
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)
}

// InitConfig reads in config file and env variables if set.
// env file can be named the ENV
// Env vars need to be uppercase with . replaced with _.
func InitConfig() *viper.Viper {
	v := viper.New()
	conf = v

	SetConfigDefaults(v)

	if os.Getenv("EXTERNAL_CONFIG") != "" {
		setupExternalConfig(v)
		return v
	}

	// Check if we're running as container on k8s
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		v.SetConfigName("config")
	} else {
		env := os.Getenv("ENV")
		if env == "" {
			env = "local"
		}
		v.SetConfigName(env)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // env vars cannot contain .
	v.AutomaticEnv()

	// Find config.yaml in ./, $CONFIGPATH/nakji/, ~/.config/nakji/, and /etc/nakji/
	v.AddConfigPath(".")
	configPath := os.Getenv("CONFIGPATH")
	if configPath == "" {
		configPath = "$HOME/.config"
	}
	v.AddConfigPath(configPath + "/nakji/")
	v.AddConfigPath("/etc/nakji/")

	// Adds root directory to config path
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	v.AddConfigPath(basepath + "/..")

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		fmt.Printf("Using defaults: %s\n", err)
		//panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	//SetFlags(v)

	if err = ValidateConfig(v); err != nil {
		panic(fmt.Errorf("%s \n", err))
	}

	InitLogging(v)

	return v
}

// setupExternalConfig uses external config file to run connectors
func setupExternalConfig(v *viper.Viper) {
	externalCfgFile := os.Getenv("EXTERNAL_CONFIG")
	v.SetConfigFile(externalCfgFile)
}

func GetAllowedBlocksBehind() uint64 {
	return conf.GetUint64("websocket.allowedBlocksBehind")
}

func GetLRUCacheSize() int {
	return conf.GetInt("websocket.lruCacheSize")
}

func GetWCReplicationFactor() int {
	return conf.GetInt("websocket.wcReplicationFactor")
}

func GetHCReplicationFactor() int {
	return conf.GetInt("websocket.hcReplicationFactor")
}

func GetMaxIdleSecond() time.Duration {
	return conf.GetDuration("websocket.maxIdleSecond")
}

func GetCoolDownSecond() time.Duration {
	return conf.GetDuration("websocket.coolDownSecond")
}

func GetPrometheusMetricsPort() int {
	return conf.GetInt("prometheus.metrics.port")
}
