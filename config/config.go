package config

import (
	"net/url"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Envar("WEB_LISTEN_ADDR").Default(":9531").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Envar("WEB_TELEMETRY_PATH").Default("/metrics").String()
	artiScrapeURI = kingpin.Flag("artifactory.scrape-uri", "URI on which to scrape JFrog Artifactory.").Envar("ARTI_SCRAPE_URI").Default("http://localhost:8081/artifactory").String()
	artiSSLVerify = kingpin.Flag("artifactory.ssl-verify", "Flag that enables SSL certificate verification for the scrape URI").Envar("ARTI_SSL_VERIFY").Default("true").Bool()
	artiTimeout   = kingpin.Flag("artifactory.timeout", "Timeout for trying to get stats from JFrog Artifactory.").Envar("ARTI_TIMEOUT").Default("5s").Duration()
)

// BasicCredentials represents Username and Password for Artifactory HTTP Basic Authentication
type BasicCredentials struct {
	Username string `required:"true" envconfig:"ARTI_USERNAME"`
	Password string `required:"true" envconfig:"ARTI_PASSWORD"`
}

// Config represents all configuration options for running the Exporter.
type Config struct {
	ListenAddress    string
	MetricsPath      string
	ArtiScrapeURI    string
	BasicCredentials *BasicCredentials
	ArtiSSLVerify    bool
	ArtiTimeout      time.Duration
	Logger           log.Logger
}

// NewConfig Creates new Artifactory exporter Config
func NewConfig() (*Config, error) {

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	var basicCredentials BasicCredentials
	err := envconfig.Process("", &basicCredentials)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(*artiScrapeURI)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "http", "https":
		return &Config{
			*listenAddress,
			*metricsPath,
			*artiScrapeURI,
			&basicCredentials,
			*artiSSLVerify,
			*artiTimeout,
			logger,
		}, nil
	}
	return nil, err
}
