package collector

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peimanja/artifactory_exporter/artifactory"
)

func (e *Exporter) exportSystem(license artifactory.LicenseInfo, ch chan<- prometheus.Metric) error {
	health, err := e.client.FetchHealth()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching system/ping",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}
	buildInfo, err := e.client.FetchBuildInfo()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching system/version",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}

	licenseType := strings.ToLower(license.Type)
	for metricName, metric := range systemMetrics {
		switch metricName {
		case "healthy":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, b2f(health.Healthy), health.NodeId)
		case "version":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, buildInfo.Version, buildInfo.Revision, buildInfo.NodeId)
		case "license":
			var validThrough float64
			timeNow := float64(time.Now().Unix())
			switch licenseType {
			case "oss", "jcr edition", "community edition for c/c++":
				validThrough = timeNow
			default:
				if validThroughTime, err := time.Parse("Jan 2, 2006", license.ValidThrough); err != nil {
					e.logger.Warn(
						"Couldn't parse Artifactory license ValidThrough",
						"err", err.Error(),
					)
					validThrough = timeNow
				} else {
					validThrough = float64(validThroughTime.Unix())
				}
			}
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, validThrough-timeNow, licenseType, license.LicensedTo, license.ValidThrough, license.NodeId)
		}
	}
	return nil
}
