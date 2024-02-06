package collector

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peimanja/artifactory_exporter/artifactory"
)

type user struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}

// map[<realm>] <count>
type realmUserCounts map[string]float64

func (e *Exporter) countUsersPerRealm(users []artifactory.User) realmUserCounts {
	e.logger.Debug("Counting users")
	usersPerRealm := realmUserCounts{}
	for _, user := range users {
		usersPerRealm[user.Realm]++
	}
	return usersPerRealm
}

func (e *Exporter) exportUsersCount(metricName string, metric *prometheus.Desc, ch chan<- prometheus.Metric) error {
	// Fetch Artifactory Users
	users, err := e.client.FetchUsers()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching security/users",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}

	usersPerRealm := e.countUsersPerRealm(users.Users)
	totalUserCount := 0
	for _, count := range usersPerRealm {
		totalUserCount += int(count)
	}

	if totalUserCount == 0 {
		e.jsonParseFailures.Inc()
		e.logger.Error("There was an issue getting users respond")
		return fmt.Errorf("There was an issue getting users respond")
	}
	for realm, count := range usersPerRealm {
		e.logger.Debug(
			"Registering metric",
			"metric", metricName,
			"realm", realm,
			"value", count,
		)
		ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, count, realm, users.NodeId)
	}
	return nil
}

type group struct {
	Name  string `json:"name"`
	Realm string `json:"uri"`
}

func (e *Exporter) exportGroups(metricName string, metric *prometheus.Desc, ch chan<- prometheus.Metric) error {
	// Fetch Artifactory groups
	groups, err := e.client.FetchGroups()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching security/users",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}

	e.logger.Debug(
		"Registering metric",
		"metric", metricName,
		"value", float64(len(groups.Groups)), // What for log as float?Int is not precise enough?
	)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, float64(len(groups.Groups)), groups.NodeId)
	return nil
}

func (e *Exporter) exportCertificates(metricName string, metric *prometheus.Desc, ch chan<- prometheus.Metric) error {
	// Fetch Artifactory certificates
	certs, err := e.client.FetchCertificates()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching system/security/certificates",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}
	if len(certs.Certificates) == 0 {
		e.logger.Debug("No certificates found")
		return nil
	}

	for _, certificate := range certs.Certificates {
		var validThrough float64
		timeNow := float64(time.Now().Unix())
		if validThroughTime, err := time.Parse(time.RFC3339, certificate.ValidUntil); err != nil {
			e.logger.Warn(
				"Couldn't parse certificate ValidThrough",
				"err", err.Error(),
			)
			validThrough = timeNow
		} else {
			validThrough = float64(validThroughTime.Unix())
		}

		alias := certificate.CertificateAlias
		issued_by := certificate.IssuedBy
		valid_until := certificate.ValidUntil
		e.logger.Debug(
			"Registering metric",
			"metric", metricName,
			"alias", alias,
			"issued_by", issued_by,
			"valid_until", valid_until,
		)
		ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, validThrough-timeNow, alias, issued_by, valid_until, certs.NodeId)
	}

	return nil
}
