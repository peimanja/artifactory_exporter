package collector

import (
	"fmt"

	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/prometheus/client_golang/prometheus"
)

type user struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}

// map[<realm>] <count>
type realmUserCounts map[string]float64

func (e *Exporter) countUsersPerRealm(users []artifactory.User) realmUserCounts {
	level.Debug(e.logger).Log("msg", "Counting users")
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
		level.Error(e.logger).Log("msg", "Couldn't scrape Artifactory when fetching security/users", "err", err)
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
		level.Error(e.logger).Log("err", "There was an issue getting users respond")
		return fmt.Errorf("There was an issue getting users respond")
	}
	for realm, count := range usersPerRealm {
		level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "realm", realm, "value", count)
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
		level.Error(e.logger).Log("msg", "Couldn't scrape Artifactory when fetching security/users", "err", err)
		e.totalAPIErrors.Inc()
		return err
	}

	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", float64(len(groups.Groups)))
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, float64(len(groups.Groups)), groups.NodeId)
	return nil
}
