package artifactory

import (
	"crypto/tls"
	"net/http"

	"github.com/go-kit/log"
	"github.com/peimanja/artifactory_exporter/config"
)

// Client represents Artifactory HTTP Client
type Client struct {
	URI             string
	authMethod      string
	cred            config.Credentials
	optionalMetrics config.OptionalMetrics
	client          *http.Client
	logger          log.Logger
}

// NewClient returns an initialized Artifactory HTTP Client.
func NewClient(conf *config.Config) *Client {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !conf.ArtiSSLVerify}}
	client := &http.Client{
		Timeout:   conf.ArtiTimeout,
		Transport: tr,
	}
	return &Client{
		URI:             conf.ArtiScrapeURI,
		authMethod:      conf.Credentials.AuthMethod,
		cred:            *conf.Credentials,
		optionalMetrics: conf.OptionalMetrics,
		client:          client,
		logger:          conf.Logger,
	}
}
