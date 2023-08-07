package artifactory

import (
	"github.com/go-kit/log/level"
)

const openMetricsEndpoint = "v1/metrics"

type OpenMetrics struct {
	PromMetrics string
	NodeId      string
}

// FetchOpenMetrics makes the API call to open metrics endpoint and returns all the open metrics
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

	level.Debug(c.logger).Log("msg", "OpenMetrics from Artifactory", "body", string(resp.Body))

	openMetrics.NodeId = resp.NodeId
	openMetrics.PromMetrics = string(resp.Body)

	return openMetrics, nil
}
