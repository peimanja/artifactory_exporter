package collector

import (
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/peimanja/artifactory_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects JFrog Artifactory stats from the given URI and
// exports them using the prometheus metrics package.
type Exporter struct {
	URI        string
	cred       config.Credentials
	authMethod string
	sslVerify  bool
	timeout    time.Duration
	mutex      sync.RWMutex

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
	logger                          log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(conf *config.Config) (*Exporter, error) {

	return &Exporter{
		URI:        conf.ArtiScrapeURI,
		cred:       *conf.Credentials,
		authMethod: conf.Credentials.AuthMethod,
		sslVerify:  conf.ArtiSSLVerify,
		timeout:    conf.ArtiTimeout,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of artifactory successful.",
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
		logger: conf.Logger,
	}, nil
}
