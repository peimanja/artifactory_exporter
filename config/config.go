package config

import (
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	l "github.com/peimanja/artifactory_exporter/logger"
)

var (
	flagLogFormat          = kingpin.Flag(l.FormatFlagName, l.FormatFlagHelp).Default(l.FormatDefault).Enum(l.FormatsAvailable...)
	flagLogLevel           = kingpin.Flag(l.LevelFlagName, l.LevelFlagHelp).Default(l.LevelDefault).Enum(l.LevelsAvailable...)
	listenAddress          = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Envar("WEB_LISTEN_ADDR").Default(":9531").String()
	metricsPath            = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Envar("WEB_TELEMETRY_PATH").Default("/metrics").String()
	artiScrapeURI          = kingpin.Flag("artifactory.scrape-uri", "URI on which to scrape JFrog Artifactory.").Envar("ARTI_SCRAPE_URI").Default("http://localhost:8081/artifactory").String()
	artiSSLVerify          = kingpin.Flag("artifactory.ssl-verify", "Flag that enables SSL certificate verification for the scrape URI").Envar("ARTI_SSL_VERIFY").Default("false").Bool()
	artiTimeout            = kingpin.Flag("artifactory.timeout", "Timeout for trying to get stats from JFrog Artifactory.").Envar("ARTI_TIMEOUT").Default("5s").Duration()
	optionalMetrics        = kingpin.Flag("optional-metric", fmt.Sprintf("optional metric to be enabled. Valid metrics are: %v", optionalMetricsList)).PlaceHolder("metric-name").Strings()
	accessFederationTarget = kingpin.Flag("access-federation-target", "URL of Jfrog Access Federation Target server. Only required if optional metric AccessFederationValidate is enabled").Envar("ACCESS_FEDERATION_TARGET").String()
	useCache               = kingpin.Flag("use-cache", "Use cache for API responses to circumvent timeouts").Envar("USE_CACHE").Default("false").Bool()
	cacheTimeout           = kingpin.Flag("cache-timeout", "Timeout for API responses to fallback to cache").Envar("CACHE_TIMEOUT").Default("30s").Duration()
	cacheTTL               = kingpin.Flag("cache-ttl", "Time to live for cached API responses").Envar("CACHE_TTL").Default("5m").Duration()
)

var optionalMetricsList = []string{"artifacts", "replication_status", "federation_status", "open_metrics", "access_federation_validate", "background_tasks"}

// Credentials represents Username and Password or API Key for
// Artifactory Authentication
type Credentials struct {
	AuthMethod  string
	Username    string `required:"false" envconfig:"ARTI_USERNAME"`
	Password    string `required:"false" envconfig:"ARTI_PASSWORD"`
	AccessToken string `required:"false" envconfig:"ARTI_ACCESS_TOKEN"`
}

// Updated OptionalMetrics struct to include YAML tags for better configuration management
type OptionalMetrics struct {
	Artifacts                bool `yaml:"artifacts"`
	ReplicationStatus        bool `yaml:"replication_status"`
	FederationStatus         bool `yaml:"federation_status"`
	OpenMetrics              bool `yaml:"open_metrics"`
	AccessFederationValidate bool `yaml:"access_federation_validate"`
	BackgroundTasks          bool `yaml:"background_tasks"`
}

// Config represents all configuration options for running the Exporter.
type Config struct {
	ListenAddress          string
	MetricsPath            string
	ArtiScrapeURI          string
	Credentials            *Credentials
	ArtiSSLVerify          bool
	ArtiTimeout            time.Duration
	UseCache               bool
	CacheTimeout           time.Duration
	CacheTTL               time.Duration
	OptionalMetrics        OptionalMetrics
	AccessFederationTarget string
	Logger                 *slog.Logger
}

// NewConfig Creates Config for Artifactory exporter
func NewConfig() (*Config, error) {

	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Info() + " " + version.BuildContext())
	kingpin.Parse()

	var credentials Credentials
	err := envconfig.Process("", &credentials)
	if err != nil {
		return nil, err
	}
	if credentials.Username != "" && credentials.Password != "" && credentials.AccessToken == "" {
		credentials.AuthMethod = "userPass"
	} else if credentials.Username == "" && credentials.Password == "" && credentials.AccessToken != "" {
		credentials.AuthMethod = "accessToken"
	} else {
		return nil, fmt.Errorf("`ARTI_USERNAME` and `ARTI_PASSWORD` or `ARTI_ACCESS_TOKEN` environment variable has to be set")
	}

	_, err = url.Parse(*artiScrapeURI)
	if err != nil {
		return nil, err
	}

	optMetrics := OptionalMetrics{}
	for _, metric := range *optionalMetrics {
		switch metric {
		case "artifacts":
			optMetrics.Artifacts = true
		case "replication_status":
			optMetrics.ReplicationStatus = true
		case "federation_status":
			optMetrics.FederationStatus = true
		case "open_metrics":
			optMetrics.OpenMetrics = true
		case "access_federation_validate":
			optMetrics.AccessFederationValidate = true
		case "background_tasks":
			optMetrics.BackgroundTasks = true
		default:
			return nil, fmt.Errorf("unknown optional metric: %s. Valid optional metrics are: %v", metric, optionalMetricsList)
		}
	}

	if *accessFederationTarget != "" {
		_, err = url.Parse(*accessFederationTarget)
		if err != nil {
			return nil, err
		}
	} else if optMetrics.AccessFederationValidate {
		return nil, fmt.Errorf("JFrog Access Federation target URL must be set if optional metric AccessFederationValidate is enabled.")
	}

	logger := l.New(
		l.Config{
			Format: *flagLogFormat,
			Level:  *flagLogLevel,
		},
	)
	return &Config{
		ListenAddress:          *listenAddress,
		MetricsPath:            *metricsPath,
		ArtiScrapeURI:          *artiScrapeURI,
		Credentials:            &credentials,
		ArtiSSLVerify:          *artiSSLVerify,
		ArtiTimeout:            *artiTimeout,
		UseCache:               *useCache,
		CacheTimeout:           *cacheTimeout,
		CacheTTL:               *cacheTTL,
		OptionalMetrics:        optMetrics,
		AccessFederationTarget: *accessFederationTarget,
		Logger:                 logger,
	}, nil

}
