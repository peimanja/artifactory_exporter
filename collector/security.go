package collector

import (
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/prometheus/client_golang/prometheus"
)

type user struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}

type usersCount struct {
	count float64
	realm string
}

func (e *Exporter) countUsers(users []artifactory.User) []usersCount {
	level.Debug(e.logger).Log("msg", "Counting users")
	userCount := []usersCount{
		{0, "saml"},
		{0, "internal"},
	}

	for _, user := range users {
		switch user.Realm {
		case "saml":
			userCount[0].count++
		case "internal":
			userCount[1].count++
		}
	}
	return userCount
}

func (e *Exporter) exportUsersCount(metricName string, metric *prometheus.Desc, ch chan<- prometheus.Metric) error {
	// Fetch Security stats
	users, err := e.c.FetchUsers()
	if err != nil {
		level.Error(e.logger).Log("msg", "Couldn't scrape Artifactory when fetching security/users", "err", err)
		return err
	}

	// Count Users
	usersCount := e.countUsers(users)

	if usersCount[0].count == 0 && usersCount[1].count == 0 {
		e.jsonParseFailures.Inc()
		level.Error(e.logger).Log("err", "There was an issue getting users respond")
		return fmt.Errorf("There was an issue getting users respond")
	}
	for _, user := range usersCount {
		level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "realm", user.realm, "value", user.count)
		ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, user.count, user.realm)
	}
	return nil
}

type group struct {
	Name  string `json:"name"`
	Realm string `json:"uri"`
}

func (e *Exporter) exportGroups(metricName string, metric *prometheus.Desc, ch chan<- prometheus.Metric) error {
	var groups []group
	level.Debug(e.logger).Log("msg", "Fetching groups stats")
	resp, err := e.fetchHTTP("security/groups")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(resp, &groups); err != nil {
		level.Error(e.logger).Log("err", "There was an issue getting groups respond")
		e.jsonParseFailures.Inc()
		return err
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", float64(len(groups)))
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, float64(len(groups)))

	return nil
}
