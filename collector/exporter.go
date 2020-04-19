package collector

import (
	"crypto/tls"
	"net/http"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/peimanja/artifactory_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

// Exporter collects JFrog Artifactory stats from the given URI and
// exports them using the prometheus metrics package.
type Exporter struct {
	c          *artifactory.Client
	URI        string
	authMethod string
	cred       config.Credentials
	client     *http.Client
	mutex      sync.RWMutex

	up                                              prometheus.Gauge
	totalScrapes, totalAPIErrors, jsonParseFailures prometheus.Counter
	logger                                          log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(conf *config.Config) (*Exporter, error) {
	c := artifactory.NewClient(conf)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !conf.ArtiSSLVerify}}
	client := &http.Client{
		Timeout:   conf.ArtiTimeout,
		Transport: tr,
	}
	return &Exporter{
		c:          c,
		URI:        conf.ArtiScrapeURI,
		cred:       *conf.Credentials,
		client:     client,
		authMethod: conf.Credentials.AuthMethod,
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
		logger: conf.Logger,
	}, nil
}
