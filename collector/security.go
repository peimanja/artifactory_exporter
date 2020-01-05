package collector

import (
	"encoding/json"

	"github.com/prometheus/client_golang/prometheus"
)

type user struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}

func (e *Exporter) fetchUsers() ([]user, error) {
	var users []user
	resp, err := fetchHTTP(e.URI, "security/users", e.bc, e.sslVerify, e.timeout)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &users); err != nil {
		e.jsonParseFailures.Inc()
		return users, err
	}
	return users, nil
}

type usersCount struct {
	count float64
	realm string
}

func (e *Exporter) countUsers(metricName string, metric *prometheus.Desc, users []user, ch chan<- prometheus.Metric) {
	userCount := []usersCount{
		{0, "saml"},
		{0, "internal"},
	}

	for _, user := range users {
		if user.Realm == "saml" {
			userCount[0].count += 1
		} else if user.Realm == "internal" {
			userCount[1].count += 1
		}
	}
	e.exportUsersCount(metricName, metric, userCount, ch)
}

func (e *Exporter) exportUsersCount(metricName string, metric *prometheus.Desc, users []usersCount, ch chan<- prometheus.Metric) {
	if users[0].count == 0 && users[1].count == 0 {
		e.jsonParseFailures.Inc()
		return
	}
	for _, user := range users {
		ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, user.count, user.realm)
	}
}

type group struct {
	Name  string `json:"name"`
	Realm string `json:"uri"`
}

func (e *Exporter) fetchGroups() ([]group, error) {
	var groups []group
	resp, err := fetchHTTP(e.URI, "security/groups", e.bc, e.sslVerify, e.timeout)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &groups); err != nil {
		e.jsonParseFailures.Inc()
		return groups, err
	}
	return groups, nil
}
