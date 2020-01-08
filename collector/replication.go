package collector

import (
	"encoding/json"
	"strings"

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
	EnableEventReplication          bool   `json:"enableEventReplication"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
	SyncStatistics                  bool   `json:"syncStatistics"`
}

func (e *Exporter) fetchReplications() ([]replication, error) {
	var replications []replication
	resp, err := fetchHTTP(e.URI, "replications", e.cred, e.authMethod, e.sslVerify, e.timeout)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &replications); err != nil {
		e.jsonParseFailures.Inc()
		return replications, err
	}
	return replications, nil
}

func (e *Exporter) exportReplications(replications []replication, ch chan<- prometheus.Metric) {
	if len(replications) == 0 {
		return
	}
	for _, replication := range replications {
		for metricName, metric := range replicationMetrics {
			switch metricName {
			case "enabled":
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, b2f(replication.Enabled), replication.RepoKey, strings.ToLower(replication.ReplicationType), replication.CronExp)

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
