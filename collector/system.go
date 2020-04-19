package collector

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) fetchHealth() (float64, error) {
	level.Debug(e.logger).Log("msg", "Fetching health stats")
	resp, err := e.fetchHTTP("system/ping")
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
	resp, err := e.fetchHTTP("system/version")
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
	resp, err := e.fetchHTTP("system/license")
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

func (e *Exporter) exportSystem(license licenseInfo, ch chan<- prometheus.Metric) error {
	healthy, err := e.fetchHealth()
	if err != nil {
		level.Error(e.logger).Log("msg", "Couldn't scrape Artifactory when fetching system/ping", "err", err)
		return err
	}
	buildInfo, err := e.fetchBuildInfo()
	if err != nil {
		level.Error(e.logger).Log("msg", "Couldn't scrape Artifactory when fetching system/version", "err", err)
		return err
	}

	licenseType := strings.ToLower(license.Type)
	for metricName, metric := range systemMetrics {
		switch metricName {
		case "healthy":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, healthy)
		case "version":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, buildInfo.Version, buildInfo.Revision)
		case "license":
			var validThrough float64
			timeNow := float64(time.Now().Unix())
			switch licenseType {
			case "oss":
				validThrough = timeNow
			default:
				if validThroughTime, err := time.Parse("Jan 2, 2006", license.ValidThrough); err != nil {
					level.Warn(e.logger).Log("msg", "Couldn't parse Artifactory license ValidThrough", "err", err)
					validThrough = timeNow
				} else {
					validThrough = float64(validThroughTime.Unix())
				}
			}
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, validThrough-timeNow, licenseType, license.LicensedTo, license.ValidThrough)
		}
	}
	return nil
}
