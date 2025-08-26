package collector

import (
	"log/slog"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/peimanja/artifactory_exporter/config"
)

// Exporter collects JFrog Artifactory stats from the given URI and
// exports them using the prometheus metrics package.
type Exporter struct {
	client                *artifactory.Client
	exporterRuntimeConfig config.ExporterRuntimeConfig
	mutex                 sync.RWMutex

	up                                              prometheus.Gauge
	totalScrapes, totalAPIErrors, jsonParseFailures prometheus.Counter
	logger                                          *slog.Logger
	backgroundTaskMetrics                           *prometheus.GaugeVec
}

// NewExporter returns an initialized Exporter.
func NewExporter(conf *config.Config) (*Exporter, error) {
	client := artifactory.NewClient(conf)

	backgroundTaskMetrics := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "artifactory_background_tasks",
			Help:      "Number of Artifactory background tasks by type and state",
		},
		[]string{"type", "state"},
	)

	return &Exporter{
		client:                client,
		exporterRuntimeConfig: *conf.ExporterRuntimeConfig,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of artifactory successful.",
		}),
		totalAPIErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_api_errors",
			Help:      "Current total API errors.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total artifactory scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_json_parse_failures",
			Help:      "Number of errors while parsing Json.",
		}),
		logger:                conf.Logger,
		backgroundTaskMetrics: backgroundTaskMetrics,
	}, nil
}
