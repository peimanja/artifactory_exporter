package collector

import (
	"strconv"

	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) exportSystem(ch chan<- prometheus.Metric) error {
	healthInfo, err := e.client.FetchHealth()
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
	licenseInfo, err := e.client.FetchLicense()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching system/license",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}
	licenseValSec, err := licenseInfo.ValidSeconds()
	if err != nil {
		e.logger.Warn(
			"Couldn't get Artifactory license validity",
			"err", err.Error(),
		) // To preserve the operation, we do nothing but log the event,
	}

	for metricName, metric := range systemMetrics {
		switch metricName {
		case "healthy":
			ch <- prometheus.MustNewConstMetric(
				metric,
				prometheus.GaugeValue,
				convArtiToPromBool(healthInfo.Healthy),
				healthInfo.NodeId,
			)
		case "version":
			ch <- prometheus.MustNewConstMetric(
				metric,
				prometheus.GaugeValue,
				1,
				buildInfo.Version,
				buildInfo.Revision,
				buildInfo.NodeId,
			)
		case "license":
			ch <- prometheus.MustNewConstMetric(
				metric,
				prometheus.GaugeValue,
				float64(licenseValSec), // Prometheus expects a float type.
				licenseInfo.TypeNormalized(),
				licenseInfo.LicensedTo,
				licenseInfo.ValidThrough,
				licenseInfo.NodeId,
			)
		}
	}
	if !licenseInfo.IsOSS() { // Some endpoints are only available commercially.
		err := e.exportAllSecurityMetrics(ch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Exporter) exportSystemHALicenses(ch chan<- prometheus.Metric) error {
	licensesInfo, err := e.client.FetchLicenses()
	if err != nil {
		// Check if this is a 404 error (endpoint not found) - common for Edge nodes
		if apiError, ok := err.(*artifactory.APIError); ok && apiError.Status() == 404 {
			e.logger.Debug(
				"HA licenses endpoint not available - likely an Edge node",
				"endpoint", "/system/licenses",
			)
			return nil // Don't treat this as an error for Edge nodes
		}
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching system/licenses",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}

	for _, licenseInfo := range licensesInfo.Licenses {
		licenseValSec, err := licenseInfo.ValidSeconds()
		if err != nil {
			e.logger.Warn(
				"Couldn't get Artifactory license validity",
				"err", err.Error(),
			) // To preserve the operation, we do nothing but log the event,
		}
		metric := systemMetrics["licenses"]
		ch <- prometheus.MustNewConstMetric(
			metric,
			prometheus.GaugeValue,
			float64(licenseValSec), // Prometheus expects a float type.
			licenseInfo.TypeNormalized(),
			licenseInfo.ValidThrough,
			licenseInfo.LicensedTo,
			licenseInfo.NodeUrl,
			licenseInfo.LicenseHash,
			strconv.FormatBool(licenseInfo.Expired),
			licenseInfo.NodeId,
		)
	}

	return nil
}
