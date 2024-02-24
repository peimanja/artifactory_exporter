package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) exportReplications(ch chan<- prometheus.Metric) error {
	// Fetch Replications stats
	replications, err := e.client.FetchReplications()
	if err != nil {
		e.logger.Error(
			"Couldn't scrape Artifactory when fetching replications",
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
		return err
	}
	if len(replications.Replications) == 0 {
		e.logger.Debug("No replications stats found")
		return nil
	}
	for _, replication := range replications.Replications {
		for metricName, metric := range replicationMetrics {
			switch metricName {
			case "enabled":
				enabled := convArtiBoolToProm(replication.Enabled)
				repo := replication.RepoKey
				rType := strings.ToLower(replication.ReplicationType)
				rURL := strings.ToLower(replication.URL)
				cronExp := replication.CronExp
				status := replication.Status
				e.logger.Debug(
					"Registering metric",
					"metric", metricName,
					"repo", replication.RepoKey,
					"type", rType,
					"url", rURL,
					"cron", cronExp,
					"status", status,
					"value", enabled,
				)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, enabled, repo, rType, rURL, cronExp, status, replications.NodeId)
			}
		}
	}
	return nil
}
