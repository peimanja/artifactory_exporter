package collector

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) exportReplications(ch chan<- prometheus.Metric) error {
	// Fetch Replications stats
	replications, err := e.client.FetchReplications()
	if err != nil {
		level.Error(e.logger).Log("msg", "Couldn't scrape Artifactory when fetching replications", "err", err)
		e.totalAPIErrors.Inc()
		return err
	}
	if len(replications.Replications) == 0 {
		level.Debug(e.logger).Log("msg", "No replications stats found")
		return nil
	}
	for _, replication := range replications.Replications {
		for metricName, metric := range replicationMetrics {
			switch metricName {
			case "enabled":
				enabled := b2f(replication.Enabled)
				repo := replication.RepoKey
				rType := strings.ToLower(replication.ReplicationType)
				rURL := strings.ToLower(replication.URL)
				cronExp := replication.CronExp
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", replication.RepoKey, "type", rType, "url", rURL, "cron", cronExp, "value", enabled)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, enabled, repo, rType, rURL, cronExp, replications.NodeId)
			}
		}
	}
	return nil
}
