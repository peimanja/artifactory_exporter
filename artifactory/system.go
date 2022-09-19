package artifactory

import (
	"encoding/json"

	"github.com/go-kit/kit/log/level"
)

const (
	pingEndpoint    = "system/ping"
	versionEndpoint = "system/version"
	licenseEndpoint = "system/license"
)

type HealthStatus struct {
	Healthy bool
	NodeId  string
}

// FetchHealth returns true if the ping endpoint returns "OK"
func (c *Client) FetchHealth() (HealthStatus, error) {
	health := HealthStatus{Healthy: false}
	level.Debug(c.logger).Log("msg", "Fetching health stats")
	resp, err := c.FetchHTTP(pingEndpoint)
	if err != nil {
		return health, err
	}
	health.NodeId = resp.NodeId
	bodyString := string(resp.Body)
	if bodyString == "OK" {
		level.Debug(c.logger).Log("msg", "System ping returned OK")
		health.Healthy = true
		return health, nil
	}
	return health, err
}

// BuildInfo represents API respond from version endpoint
type BuildInfo struct {
	Version  string   `json:"version"`
	Revision string   `json:"revision"`
	Addons   []string `json:"addons"`
	License  string   `json:"license"`
	NodeId   string
}

// FetchBuildInfo makes the API call to version endpoint and returns BuildInfo
func (c *Client) FetchBuildInfo() (BuildInfo, error) {
	var buildInfo BuildInfo
	level.Debug(c.logger).Log("msg", "Fetching build stats")
	resp, err := c.FetchHTTP(versionEndpoint)
	if err != nil {
		return buildInfo, err
	}
	buildInfo.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &buildInfo); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal buildInfo respond")
		return buildInfo, &UnmarshalError{
			message:  err.Error(),
			endpoint: versionEndpoint,
		}
	}
	return buildInfo, nil
}

// LicenseInfo represents API respond from license endpoint
type LicenseInfo struct {
	Type         string `json:"type"`
	ValidThrough string `json:"validThrough"`
	LicensedTo   string `json:"licensedTo"`
	NodeId       string
}

// FetchLicense makes the API call to license endpoint and returns LicenseInfo
func (c *Client) FetchLicense() (LicenseInfo, error) {
	var licenseInfo LicenseInfo
	level.Debug(c.logger).Log("msg", "Fetching license stats")
	resp, err := c.FetchHTTP(licenseEndpoint)
	if err != nil {
		return licenseInfo, err
	}
	licenseInfo.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &licenseInfo); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal licenseInfo respond")
		return licenseInfo, &UnmarshalError{
			message:  err.Error(),
			endpoint: licenseEndpoint,
		}
	}
	return licenseInfo, nil
}
