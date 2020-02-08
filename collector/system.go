package collector

import (
	"encoding/json"

	"github.com/go-kit/kit/log/level"
)

func (e *Exporter) fetchHealth() (float64, error) {
	level.Debug(e.logger).Log("msg", "Fetching health stats")
	resp, err := e.fetchHTTP(e.URI, "system/ping", e.cred, e.authMethod, e.sslVerify, e.timeout)
	if err != nil {
		return 0, err
	}
	bodyString := string(resp)
	if bodyString == "OK" {
		level.Debug(e.logger).Log("msg", "System ping returned OK")
		return 1, nil
	}
	return 0, err
}

type buildInfo struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
}

func (e *Exporter) fetchBuildInfo() (buildInfo, error) {
	var buildInfo buildInfo
	level.Debug(e.logger).Log("msg", "Fetching build stats")
	resp, err := e.fetchHTTP(e.URI, "system/version", e.cred, e.authMethod, e.sslVerify, e.timeout)
	if err != nil {
		return buildInfo, err
	}
	if err := json.Unmarshal(resp, &buildInfo); err != nil {
		level.Debug(e.logger).Log("msg", "There was an issue getting builds respond")
		e.jsonParseFailures.Inc()
		return buildInfo, err
	}
	return buildInfo, nil
}

type licenseInfo struct {
	Type         string `json:"type"`
	ValidThrough string `json:"validThrough"`
	LicensedTo   string `json:"licensedTo"`
}

func (e *Exporter) fetchLicense() (licenseInfo, error) {
	var licenseInfo licenseInfo
	level.Debug(e.logger).Log("msg", "Fetching license stats")
	resp, err := e.fetchHTTP(e.URI, "system/license", e.cred, e.authMethod, e.sslVerify, e.timeout)
	if err != nil {
		return licenseInfo, err
	}
	if err := json.Unmarshal(resp, &licenseInfo); err != nil {
		level.Debug(e.logger).Log("msg", "There was an issue getting license respond")
		e.jsonParseFailures.Inc()
		return licenseInfo, err
	}
	return licenseInfo, nil
}
