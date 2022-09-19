package collector

import (
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/log/level"
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
	level.Debug(e.logger).Log("msg", "Finding all artifacts", "period", period, "queryType", queryType)
	switch queryType {
	case "created":
		query = fmt.Sprintf("items.find({\"modified\" : {\"$last\" : \"%s\"}}).include(\"name\", \"repo\")", period)
	case "downloaded":
		query = fmt.Sprintf("items.find({\"stat.downloaded\" : {\"$last\" : \"%s\"}}).include(\"name\", \"repo\")", period)
	default:
		level.Error(e.logger).Log("err", "Query Type is not supported", "query", queryType)
		return artifacts, fmt.Errorf("Query Type is not supported: %s", queryType)
	}
	resp, err := e.client.QueryAQL([]byte(query))
	if err != nil {
		e.totalAPIErrors.Inc()
		return artifacts, err
	}
	artifacts.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &artifacts); err != nil {
		level.Warn(e.logger).Log("msg", "There was an error when trying to unmarshal AQL response", "queryType", queryType, "period", period, "error", err)
		e.jsonParseFailures.Inc()
		return artifacts, err
	}
	return artifacts, err
}

func (e *Exporter) getTotalArtifacts(r []repoSummary) ([]repoSummary, error) {
	created1m, err := e.findArtifacts("1minutes", "created")
	if err != nil {
		return nil, err
	}
	created5m, err := e.findArtifacts("5minutes", "created")
	if err != nil {
		return nil, err
	}
	created15m, err := e.findArtifacts("15minutes", "created")
	if err != nil {
		return nil, err
	}
	downloaded1m, err := e.findArtifacts("1minutes", "downloaded")
	if err != nil {
		return nil, err
	}
	downloaded5m, err := e.findArtifacts("5minutes", "downloaded")
	if err != nil {
		return nil, err
	}
	downloaded15m, err := e.findArtifacts("15minutes", "downloaded")
	if err != nil {
		return nil, err
	}

	repoSummaries := r
	for i := range repoSummaries {
		for _, k := range created1m.Results {
			repoSummaries[i].NodeId = created1m.NodeId
			if repoSummaries[i].Name == k.Repo {
				repoSummaries[i].TotalCreate1m++
			}
		}
		for _, k := range created5m.Results {
			repoSummaries[i].NodeId = created1m.NodeId
			if repoSummaries[i].Name == k.Repo {
				repoSummaries[i].TotalCreated5m++
			}
		}
		for _, k := range created15m.Results {
			repoSummaries[i].NodeId = created1m.NodeId
			if repoSummaries[i].Name == k.Repo {
				repoSummaries[i].TotalCreated15m++
			}
		}
		for _, k := range downloaded1m.Results {
			repoSummaries[i].NodeId = created1m.NodeId
			if repoSummaries[i].Name == k.Repo {
				repoSummaries[i].TotalDownloaded1m++
			}
		}
		for _, k := range downloaded5m.Results {
			repoSummaries[i].NodeId = created1m.NodeId
			if repoSummaries[i].Name == k.Repo {
				repoSummaries[i].TotalDownloaded5m++
			}
		}
		for _, k := range downloaded15m.Results {
			repoSummaries[i].NodeId = created1m.NodeId
			if repoSummaries[i].Name == k.Repo {
				repoSummaries[i].TotalDownloaded15m++
			}
		}
	}
	return repoSummaries, nil
}

func (e *Exporter) exportArtifacts(repoSummaries []repoSummary, ch chan<- prometheus.Metric) {
	for _, repoSummary := range repoSummaries {
		for metricName, metric := range artifactsMetrics {
			switch metricName {
			case "created1m":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.TotalCreate1m)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.TotalCreate1m, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "created5m":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.TotalCreated5m)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.TotalCreated5m, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "created15m":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.TotalCreated15m)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.TotalCreated15m, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "downloaded1m":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.TotalDownloaded1m)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.TotalDownloaded1m, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "downloaded5m":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.TotalDownloaded5m)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.TotalDownloaded5m, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "downloaded15m":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.TotalDownloaded15m)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.TotalDownloaded15m, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			}
		}
	}
}
