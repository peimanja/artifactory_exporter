package collector

import (
	"encoding/json"
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type replication struct {
	ReplicationType                 string `json:"replicationType"`
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	PathPrefix                      string `json:"pathPrefix"`
	RepoKey                         string `json:"repoKey"`
	URL                             string `json:"url"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
	SyncStatistics                  bool   `json:"syncStatistics"`
}

func (e *Exporter) fetchReplications() ([]replication, error) {
	var replications []replication
	level.Debug(e.logger).Log("msg", "Fetching replications stats")
	resp, err := e.fetchHTTP(e.URI, "replications", e.cred, e.authMethod, e.sslVerify, e.timeout)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &replications); err != nil {
		level.Debug(e.logger).Log("msg", "There was an issue getting replication respond")
		e.jsonParseFailures.Inc()
		return replications, err
	}
	return replications, nil
}

func (e *Exporter) exportReplications(replications []replication, ch chan<- prometheus.Metric) {
	if len(replications) == 0 {
		level.Debug(e.logger).Log("msg", "No replications stats found")
		return
	}
	for _, replication := range replications {
		for metricName, metric := range replicationMetrics {
			switch metricName {
			case "enabled":
				enabled := b2f(replication.Enabled)
				repo := replication.RepoKey
				rType := strings.ToLower(replication.ReplicationType)
				rURL := strings.ToLower(replication.URL)
				cronExp := replication.CronExp
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", replication.RepoKey, "type", rType, "url", rURL, "cron", cronExp, "value", enabled)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, enabled, repo, rType, rURL, cronExp)
			}
		}
	}
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
