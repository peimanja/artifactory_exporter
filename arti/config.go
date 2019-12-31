package arti

import (
	"net"
	"net/url"

	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress      = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Envar("WEB_LISTEN_ADDR").Default(":9531").String()
	metricsPath        = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Envar("WEB_TELEMETRY_PATH").Default("/metrics").String()
	artiUser           = kingpin.Flag("artifactory.user", "User to access Artifactory.").Required().Envar("ARTI_USER").String()
	artiScrapeURI      = kingpin.Flag("artifactory.scrape-uri", "URI on which to scrape Artifactory.").Envar("ARTI_SCRAPE_URI").Default("http://localhost:8081/artifactory").String()
	artiScrapeInterval = kingpin.Flag("artifactory.scrape-interval", "How often to scrape Artifactory in secoonds.").Envar("ARTI_SCRAPE_INTERVAL").Default("30").Int64()
	debug              = kingpin.Flag("exporter.debug", "Enable debug mode.").Envar("DEBUG").Default("false").Bool()
)

type APIClientConfig struct {
	Url  string
	User string
	Pass string `required:"true" envconfig:"ARTI_PASSWORD"`
}

type Config struct {
	ListenAddress      string
	MetricsPath        string
	ArtiScrapeInterval int64
	Debug              bool
	ApiConfig          *APIClientConfig
}

func NewConfig() (*Config, error) {

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var apiConfig APIClientConfig
	err := envconfig.Process("", &apiConfig)
	if err != nil {
		log.Debugln(err)
		return &Config{}, err
	}

	u, err := url.Parse(*artiScrapeURI)
	if err != nil {
		log.Debugln(err)
		return &Config{}, err
	}
	_, err = net.LookupHost(u.Hostname())
	if err != nil {
		log.Debugln(err)
		return &Config{}, err
	}

	apiConfig.Url = *artiScrapeURI
	apiConfig.User = *artiUser

	return &Config{
		*listenAddress,
		*metricsPath,
		*artiScrapeInterval,
		*debug,
		&apiConfig,
	}, nil

}
