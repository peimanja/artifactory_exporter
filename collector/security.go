package collector

import (
	"fmt"

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
