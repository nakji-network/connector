package monitor

import (
	"fmt"
	"net/http"

	"github.com/nakji-network/connector/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	port int
)

var (
	kafkaLastWriteTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "connector",
			Name:      "kafka_last_write_time",
			Help:      "connectors last time write to kafka",
		},
		[]string{"connector"},
	)

	connectorName = ""
)

func init() {
	prometheus.MustRegister(kafkaLastWriteTime)
}

// StartMonitor initiates a monitor and uses connector's name as the label value,
// and exposes an endpoint with default path: /metrics, port: 9999
func StartMonitor(name string) {
	log.Info().
		Str("name", name).
		Msg("Starting monitoring")

	connectorName = name
	port = config.GetPrometheusMetricsPort()
	if port == 0 {
		log.Error().Int("port", port).Msg("Prometheus /metrics port undefined")
		return
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(fmt.Sprint(":", port), nil)
		if err != nil {
			log.Error().Err(err).Msg("failed to start Prometheus")
			return
		}
	}()
}

// SetMetricsForKafkaLastWriteTime sets the Gauge to the current Unix time in seconds
// when connectors write to kafka.
func SetMetricsForKafkaLastWriteTime() {
	if connectorName == "" {
		log.Warn().Msg("connector name not defined for Prometheus")
		return
	}
	kafkaLastWriteTime.WithLabelValues(connectorName).SetToCurrentTime()
}
