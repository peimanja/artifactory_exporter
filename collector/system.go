package collector

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peimanja/artifactory_exporter/artifactory"
)

var afOSSLicenseTypes = []string{
	`oss`,
	`jcr edition`,
	`community edition for c/c++`,
}

func collectLicense(e *Exporter, ch chan<- prometheus.Metric) (artifactory.LicenseInfo, error) {
	retErr := func(err error) (artifactory.LicenseInfo, error) {
		return artifactory.LicenseInfo{}, err
	}

	license, err := e.client.FetchLicense()
	if err != nil {
		return retErr(err)
	}

	if !license.IsOSS() { // Some endpoints are only available commercially.
		for metricName, metric := range securityMetrics {
			switch metricName {
			case "users":
				err := e.exportUsersCount(metricName, metric, ch)
				if err != nil {
					return retErr(err)
				}
			case "groups":
				err := e.exportGroups(metricName, metric, ch)
				if err != nil {
					return retErr(err)
				}
			case "certificates":
				err := e.exportCertificates(metricName, metric, ch)
				if err != nil {
					return retErr(err)
				}
			}
		}
		if err := e.exportReplications(ch); err != nil {
			return retErr(err)
		}
	}

	licenseValidSeconds := func() int64 {
		if license.IsOSS() {
			return 0
		}
		validThroughTime, err := time.Parse("Jan 2, 2006", license.ValidThrough)
		if err != nil {
			e.logger.Warn(
				"Couldn't parse Artifactory license ValidThrough",
				"err", err.Error(),
			)
			return 0 // We deliberately ignore the error in order to maintain continuity.
		}
		validThroughEpoch := validThroughTime.Unix()
		timeNowEpoch := time.Now().Unix()
		return validThroughEpoch - timeNowEpoch
	}
	license.ValidSeconds = licenseValidSeconds()
	return license, nil
}

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
	licenseInfo, err := collectLicense(e, ch)
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching system/license",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}

	for metricName, metric := range systemMetrics {
		switch metricName {
		case "healthy":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, convArtiToPromBool(healthInfo.Healthy), healthInfo.NodeId)
		case "version":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, buildInfo.Version, buildInfo.Revision, buildInfo.NodeId)
		case "license":
			ch <- prometheus.MustNewConstMetric(
				metric,
				prometheus.GaugeValue,
				float64(licenseInfo.ValidSeconds), //float
				licenseInfo.NormalizedLicenseType(),
				licenseInfo.LicensedTo,
				licenseInfo.ValidThrough,
				licenseInfo.NodeId,
			)
		}
	}
	return nil
}
