package artifactory

import (
	"encoding/json"
	"slices"
	"strings"
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
	c.logger.Debug("Fetching health stats")
	resp, err := c.FetchHTTP(pingEndpoint)
	if err != nil {
		return health, err
	}
	health.NodeId = resp.NodeId
	bodyString := string(resp.Body)
	if bodyString == "OK" {
		c.logger.Debug("System ping returned OK")
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
	c.logger.Debug("Fetching build stats")
	resp, err := c.FetchHTTP(versionEndpoint)
	if err != nil {
		return buildInfo, err
	}
	buildInfo.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &buildInfo); err != nil {
		c.logger.Error("There was an issue when try to unmarshal buildInfo respond")
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
	ValidSeconds int64 // It will be calculated in the ‘collector’ package.
}

func (l LicenseInfo) NormalizedLicenseType() string {
	return strings.ToLower(l.Type)
}

func (l LicenseInfo) IsOSS() bool {
	var afOSSLicenseTypes = []string{
		`community edition for c/c++`,
		`jcr edition`,
		`oss`,
	}
	return slices.Contains(
		afOSSLicenseTypes,
		l.NormalizedLicenseType(),
	)
}

// FetchLicense makes the API call to license endpoint and returns LicenseInfo
func (c *Client) FetchLicense() (LicenseInfo, error) {
	var licenseInfo LicenseInfo
	c.logger.Debug("Fetching license stats")
	resp, err := c.FetchHTTP(licenseEndpoint)
	if err != nil {
		return licenseInfo, err
	}
	licenseInfo.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &licenseInfo); err != nil {
		c.logger.Error("There was an issue when try to unmarshal licenseInfo respond")
		return licenseInfo, &UnmarshalError{
			message:  err.Error(),
			endpoint: licenseEndpoint,
		}
	}
	return licenseInfo, nil
}
