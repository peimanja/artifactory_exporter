package collector

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/prometheus/client_golang/prometheus"
)

const calculateValueError = "There was an issue calculating the value"

func (e *Exporter) exportCount(metricName string, metric *prometheus.Desc, count string, nodeId string, ch chan<- prometheus.Metric) {
	if count == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.removeCommas(count)
	if err != nil {
		e.jsonParseFailures.Inc()
		level.Error(e.logger).Log("msg", calculateValueError, "metric", metricName, "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", value)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value, nodeId)
}

func (e *Exporter) exportSize(metricName string, metric *prometheus.Desc, size string, nodeId string, ch chan<- prometheus.Metric) {
	if size == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.bytesConverter(size)
	if err != nil {
		e.jsonParseFailures.Inc()
		level.Error(e.logger).Log("msg", calculateValueError, "metric", metricName, "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", value)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value, nodeId)
}

func (e *Exporter) exportFilestore(metricName string, metric *prometheus.Desc, size string, fileStoreType string, fileStoreDir string, nodeId string, ch chan<- prometheus.Metric) {
	if size == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.bytesConverter(size)
	if err != nil {
		e.jsonParseFailures.Inc()
		level.Debug(e.logger).Log("msg", calculateValueError, "metric", metricName, "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", value)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value, fileStoreType, fileStoreDir, nodeId)
}

type repoSummary struct {
	Name               string
	Type               string
	FoldersCount       float64
	FilesCount         float64
	UsedSpace          float64
	ItemsCount         float64
	PackageType        string
	Percentage         float64
	TotalCreate1m      float64
	TotalCreated5m     float64
	TotalCreated15m    float64
	TotalDownloaded1m  float64
	TotalDownloaded5m  float64
	TotalDownloaded15m float64
	NodeId             string
}

func (e *Exporter) extractRepo(storageInfo artifactory.StorageInfo) ([]repoSummary, error) {
	var err error
	rs := repoSummary{}
	repoSummaryList := []repoSummary{}
	level.Debug(e.logger).Log("msg", "Extracting repo summaries")
	for _, repo := range storageInfo.RepositoriesSummaryList {
		if repo.RepoKey == "TOTAL" {
			continue
		}
		rs.Name = repo.RepoKey
		rs.Type = strings.ToLower(repo.RepoType)
		rs.FoldersCount = float64(repo.FoldersCount)
		rs.FilesCount = float64(repo.FilesCount)
		rs.ItemsCount = float64(repo.ItemsCount)
		rs.PackageType = strings.ToLower(repo.PackageType)
		rs.UsedSpace, err = e.bytesConverter(repo.UsedSpace)
		if err != nil {
			level.Debug(e.logger).Log("msg", "There was an issue parsing repo UsedSpace", "repo", repo.RepoKey, "err", err)
			e.jsonParseFailures.Inc()
			return repoSummaryList, err
		}
		if repo.Percentage == "N/A" {
			rs.Percentage = 0
		} else {
			rs.Percentage, err = e.removeCommas(repo.Percentage)
			if err != nil {
				level.Debug(e.logger).Log("msg", "There was an issue parsing repo Percentage", "repo", repo.RepoKey, "err", err)
				e.jsonParseFailures.Inc()
				return repoSummaryList, err
			}
		}
		repoSummaryList = append(repoSummaryList, rs)
	}
	return repoSummaryList, err
}

func (e *Exporter) exportRepo(repoSummaries []repoSummary, ch chan<- prometheus.Metric) {
	for _, repoSummary := range repoSummaries {
		for metricName, metric := range storageMetrics {
			switch metricName {
			case "repoUsed":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.UsedSpace)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.UsedSpace, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoFolders":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.FoldersCount)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.FoldersCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoItems":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.ItemsCount)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.ItemsCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoFiles":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.FilesCount)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.FilesCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoPercentage":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.Percentage)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.Percentage, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			}
		}
	}
}

func (e *Exporter) exportStorage(storageInfo artifactory.StorageInfo, ch chan<- prometheus.Metric) {
	fileStoreType := strings.ToLower(storageInfo.FileStoreSummary.StorageType)
	fileStoreDir := storageInfo.FileStoreSummary.StorageDirectory
	for metricName, metric := range storageMetrics {
		switch metricName {
		case "artifacts":
			e.exportCount(metricName, metric, storageInfo.BinariesSummary.ArtifactsCount, storageInfo.NodeId, ch)
		case "artifactsSize":
			e.exportSize(metricName, metric, storageInfo.BinariesSummary.ArtifactsSize, storageInfo.NodeId, ch)
		case "binaries":
			e.exportCount(metricName, metric, storageInfo.BinariesSummary.BinariesCount, storageInfo.NodeId, ch)
		case "binariesSize":
			e.exportSize(metricName, metric, storageInfo.BinariesSummary.BinariesSize, storageInfo.NodeId, ch)
		case "filestore":
			e.exportFilestore(metricName, metric, storageInfo.FileStoreSummary.TotalSpace, fileStoreType, fileStoreDir, storageInfo.NodeId, ch)
		case "filestoreUsed":
			e.exportFilestore(metricName, metric, storageInfo.FileStoreSummary.UsedSpace, fileStoreType, fileStoreDir, storageInfo.NodeId, ch)
		case "filestoreFree":
			e.exportFilestore(metricName, metric, storageInfo.FileStoreSummary.FreeSpace, fileStoreType, fileStoreDir, storageInfo.NodeId, ch)
		}
	}
}
