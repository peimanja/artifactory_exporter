package artifactory

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/peimanja/artifactory_exporter/config"
)

// Client represents Artifactory HTTP Client
type Client struct {
	URI                    string
	authMethod             string
	cred                   config.Credentials
	optionalMetrics        config.OptionalMetrics
	accessFederationTarget string
	client                 *http.Client
	logger                 *slog.Logger
}

// NewClient returns an initialized Artifactory HTTP Client.
func NewClient(conf *config.Config) *Client {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !conf.ArtiSSLVerify}}
	client := &http.Client{
		Timeout:   conf.ArtiTimeout,
		Transport: tr,
	}
	return &Client{
		URI:                    conf.ArtiScrapeURI,
		authMethod:             conf.Credentials.AuthMethod,
		cred:                   *conf.Credentials,
		optionalMetrics:        conf.OptionalMetrics,
		accessFederationTarget: conf.AccessFederationTarget,
		client:                 client,
		logger:                 conf.Logger,
	}
}

func (c *Client) GetAccessFederationTarget() string {
	return c.accessFederationTarget
}

func (c *Client) FetchHTTPWithContext(ctx context.Context, endpoint string) (*ApiResponse, error) {
	fullURL := fmt.Sprintf("%s/api/%s", c.URI, endpoint)
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &ApiResponse{
		Body:       body,
		NodeId: resp.Header.Get("X-Artifactory-Node-Id"),
	}, nil
}
