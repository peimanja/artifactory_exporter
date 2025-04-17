package artifactory

import (
	"context"
	"crypto/tls"
	"encoding/json"
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
	OptionalMetrics        config.OptionalMetrics
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
		OptionalMetrics:        conf.OptionalMetrics,
		accessFederationTarget: conf.AccessFederationTarget,
		client:                 client,
		logger:                 conf.Logger,
	}
}

func (c *Client) GetAccessFederationTarget() string {
	return c.accessFederationTarget
}

// FetchHTTPWithContext makes a GET request to the Artifactory API with a context-aware timeout.
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
		Body:   body,
		NodeId: resp.Header.Get("X-Artifactory-Node-Id"),
	}, nil
}

// FetchBackgroundTasks makes the API call to the background tasks endpoint and returns a list of tasks
func (c *Client) FetchBackgroundTasks() ([]BackgroundTask, error) {
	const backgroundTasksEndpoint = "tasks"

	var tasksResponse struct {
		Tasks []BackgroundTask `json:"tasks"`
	}

	c.logger.Debug("Fetching background tasks")
	resp, err := c.FetchHTTP(backgroundTasksEndpoint)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resp.Body, &tasksResponse); err != nil {
		c.logger.Error("There was an issue when trying to unmarshal background tasks response")
		return nil, &UnmarshalError{
			message:  err.Error(),
			endpoint: backgroundTasksEndpoint,
		}
	}

	return tasksResponse.Tasks, nil
}

type BackgroundTask struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	State       string `json:"state"`
	Description string `json:"description"`
	NodeID      string `json:"nodeId"`
}
