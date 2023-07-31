package artifactory

import (
	"github.com/go-kit/log/level"
)

const openMetricsEndpoint = "v1/metrics"

type OpenMetrics struct {
	Metric string
	NodeId string
}

// FetchReplications makes the API call to replication endpoint and returns []Replication
func (c *Client) FetchOpenMetrics() (OpenMetrics, error) {
	var openMetrics OpenMetrics
	level.Debug(c.logger).Log("msg", "Fetching openMetrics")
	resp, err := c.FetchHTTP(openMetricsEndpoint)
	if err != nil {
		if err.(*APIError).status == 404 {
			return openMetrics, nil
		}
		return openMetrics, err
	}
	openMetrics.NodeId = resp.NodeId
	openMetrics.Metric = string(resp.Body)
	level.Debug(c.logger).Log("msg", "OpenMetrics from Artifactory", "body", string(resp.Body))

	return openMetrics, nil
}
