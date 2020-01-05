package collector

import "encoding/json"

func (e *Exporter) fetchHealth() (float64, error) {
	resp, err := fetchHTTP(e.URI, "system/ping", e.bc, e.sslVerify, e.timeout)
	if err != nil {
		return 0, err
	}
	bodyString := string(resp)
	if bodyString == "OK" {
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
	resp, err := fetchHTTP(e.URI, "system/version", e.bc, e.sslVerify, e.timeout)
	if err != nil {
		return buildInfo, err
	}
	if err := json.Unmarshal(resp, &buildInfo); err != nil {
		e.jsonParseFailures.Inc()
		return buildInfo, err
	}
	return buildInfo, nil
}
