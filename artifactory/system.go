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

// FetchHealth returns true if the ping endpoint returns "OK"
func (c *Client) FetchHealth() (bool, error) {
	level.Debug(c.logger).Log("msg", "Fetching health stats")
	resp, err := c.fetchHTTP(pingEndpoint)
	if err != nil {
		return false, err
	}
	bodyString := string(resp)
	if bodyString == "OK" {
		level.Debug(c.logger).Log("msg", "System ping returned OK")
		return true, nil
	}
	return false, err
}

// BuildInfo represents API respond from version endpoint
type BuildInfo struct {
	Version  string   `json:"version"`
	Revision string   `json:"revision"`
	Addons   []string `json:"addons"`
	License  string   `json:"license"`
}

// FetchBuildInfo makes the API call to version endpoint and returns BuildInfo
func (c *Client) FetchBuildInfo() (BuildInfo, error) {
	var buildInfo BuildInfo
	level.Debug(c.logger).Log("msg", "Fetching build stats")
	resp, err := c.fetchHTTP(versionEndpoint)
	if err != nil {
		return buildInfo, err
	}
	if err := json.Unmarshal(resp, &buildInfo); err != nil {
		level.Debug(c.logger).Log("msg", "There was an issue getting builds respond")
		return buildInfo, err
	}
	return buildInfo, nil
}

// LicenseInfo represents API respond from license endpoint
type LicenseInfo struct {
	Type         string `json:"type"`
	ValidThrough string `json:"validThrough"`
	LicensedTo   string `json:"licensedTo"`
}

// FetchLicense makes the API call to license endpoint and returns LicenseInfo
func (c *Client) FetchLicense() (LicenseInfo, error) {
	var licenseInfo LicenseInfo
	level.Debug(c.logger).Log("msg", "Fetching license stats")
	resp, err := c.fetchHTTP(licenseEndpoint)
	if err != nil {
		return licenseInfo, err
	}
	if err := json.Unmarshal(resp, &licenseInfo); err != nil {
		level.Debug(c.logger).Log("msg", "There was an issue getting license respond")
		return licenseInfo, err
	}
	return licenseInfo, nil
}
