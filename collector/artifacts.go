package collector

import (
	"encoding/json"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type artifact struct {
	Repo string `json:"repo,omitempty"`
	Name string `json:"name,omitempty"`
}

type artifactQueryResult struct {
	Results []artifact `json:"results,omitempty"`
	NodeId  string
}

func (e *Exporter) findArtifacts(period string, queryType string) (artifactQueryResult, error) {
	var query string
	artifacts := artifactQueryResult{}
	e.logger.Debug(
		"Finding all artifacts",
		"period", period,
		"queryType", queryType,
	)
	switch queryType {
	case "created":
		query = fmt.Sprintf("items.find({\"modified\" : {\"$last\" : \"%s\"}}).include(\"name\", \"repo\")", period)
	case "downloaded":
		query = fmt.Sprintf("items.find({\"stat.downloaded\" : {\"$last\" : \"%s\"}}).include(\"name\", \"repo\")", period)
	default:
		e.logger.Error(
			"Query Type is not supported",
			"query", queryType,
		)
		return artifacts, fmt.Errorf("Query Type is not supported: %s", queryType)
	}
	resp, err := e.client.QueryAQL([]byte(query))
	if err != nil {
		e.totalAPIErrors.Inc()
		return artifacts, err
	}
	artifacts.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &artifacts); err != nil {
		e.logger.Warn(
			"There was an error when trying to unmarshal AQL response",
			"queryType", queryType,
			"period", period,
			"error", err.Error(),
		)
		e.jsonParseFailures.Inc()
		return artifacts, err
	}
	return artifacts, err
}

func (e *Exporter) getTotalArtifacts(r []repoSummary) ([]repoSummary, error) {
	repoSummaries := r

	timeIntervals := e.exporterRuntimeConfig.ArtifactsTimeIntervals

	groupedRepoSummary := make(map[string]*repoSummary, len(repoSummaries))
	for rep_i, repo := range repoSummaries {
		repoSummaries[rep_i].RepoArtifactsSummary = make([]RepoArtifactsSummary, len(timeIntervals))
		// Fill the slice directly
		for interval_i, timeInterval := range timeIntervals {
			repoSummaries[rep_i].RepoArtifactsSummary[interval_i] = RepoArtifactsSummary{period: timeInterval.ShortPeriod}
		}
		groupedRepoSummary[repo.Name] = &repoSummaries[rep_i]
	}

	for interval_i, timeInterval := range timeIntervals {
		created, err := e.findArtifacts(timeInterval.Period, "created")
		if err != nil {
			return nil, err
		}
		downloaded, err := e.findArtifacts(timeInterval.Period, "downloaded")
		if err != nil {
			return nil, err
		}

		for _, item := range created.Results {
			groupedRepoSummary[item.Repo].RepoArtifactsSummary[interval_i].TotalCreated++
		}
		for _, item := range downloaded.Results {
			groupedRepoSummary[item.Repo].RepoArtifactsSummary[interval_i].TotalDownloaded++
		}
	}

	return repoSummaries, nil
}

func (e *Exporter) exportArtifacts(repoSummaries []repoSummary, ch chan<- prometheus.Metric) {
	for _, repoSummary := range repoSummaries {
		for _, repoArtifactsSummary := range repoSummary.RepoArtifactsSummary {
			createdMetricName := fmt.Sprintf("created_%s", repoArtifactsSummary.period)
			downloadedMetricName := fmt.Sprintf("downloaded_%s", repoArtifactsSummary.period)

			e.logger.Debug(
				logDbgMsgRegMetric,
				"metric", createdMetricName,
				"repo", repoSummary.Name,
				"type", repoSummary.Type,
				"package_type", repoSummary.PackageType,
				"value", repoArtifactsSummary.TotalCreated,
			)
			createdMetric := artifactsMetrics[createdMetricName]
			ch <- prometheus.MustNewConstMetric(createdMetric, prometheus.GaugeValue, repoArtifactsSummary.TotalCreated, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)

			e.logger.Debug(
				logDbgMsgRegMetric,
				"metric", downloadedMetricName,
				"repo", repoSummary.Name,
				"type", repoSummary.Type,
				"package_type", repoSummary.PackageType,
				"value", repoArtifactsSummary.TotalDownloaded,
			)
			downloadedMetric := artifactsMetrics[downloadedMetricName]
			ch <- prometheus.MustNewConstMetric(downloadedMetric, prometheus.GaugeValue, repoArtifactsSummary.TotalDownloaded, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
		}
	}
}
