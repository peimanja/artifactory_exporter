package collector

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) exportCount(metricName string, metric *prometheus.Desc, count string, ch chan<- prometheus.Metric) {
	if count == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.removeCommas(count)
	if err != nil {
		e.jsonParseFailures.Inc()
		level.Error(e.logger).Log("msg", "There was an issue calculating the value", "metric", metricName, "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", value)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value)
}

func (e *Exporter) exportSize(metricName string, metric *prometheus.Desc, size string, ch chan<- prometheus.Metric) {
	if size == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.bytesConverter(size)
	if err != nil {
		e.jsonParseFailures.Inc()
		level.Error(e.logger).Log("msg", "There was an issue calculating the value", "metric", metricName, "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "value", value)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value)
}

func (e *Exporter) exportFilestore(metricName string, metric *prometheus.Desc, size string, fileStoreType string, fileStoreDir string, ch chan<- prometheus.Metric) {
	if size == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.bytesConverter(size)
	if err != nil {
		e.jsonParseFailures.Inc()
		level.Debug(e.logger).Log("msg", "There was an issue calculating the value", "metric", metricName, "err", err)
		return
	}
	level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "type", fileStoreType, "directory", fileStoreDir, "value", value)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value, fileStoreType, fileStoreDir)
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
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.UsedSpace, repoSummary.Name, repoSummary.Type, repoSummary.PackageType)
			case "repoFolders":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.FoldersCount)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.FoldersCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType)
			case "repoItems":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.ItemsCount)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.ItemsCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType)
			case "repoFiles":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.FilesCount)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.FilesCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType)
			case "repoPercentage":
				level.Debug(e.logger).Log("msg", "Registering metric", "metric", metricName, "repo", repoSummary.Name, "type", repoSummary.Type, "package_type", repoSummary.PackageType, "value", repoSummary.Percentage)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.Percentage, repoSummary.Name, repoSummary.Type, repoSummary.PackageType)
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
			e.exportCount(metricName, metric, storageInfo.BinariesSummary.ArtifactsCount, ch)
		case "artifactsSize":
			e.exportSize(metricName, metric, storageInfo.BinariesSummary.ArtifactsSize, ch)
		case "binaries":
			e.exportCount(metricName, metric, storageInfo.BinariesSummary.BinariesCount, ch)
		case "binariesSize":
			e.exportSize(metricName, metric, storageInfo.BinariesSummary.BinariesSize, ch)
		case "filestore":
			e.exportFilestore(metricName, metric, storageInfo.FileStoreSummary.TotalSpace, fileStoreType, fileStoreDir, ch)
		case "filestoreUsed":
			e.exportFilestore(metricName, metric, storageInfo.FileStoreSummary.UsedSpace, fileStoreType, fileStoreDir, ch)
		case "filestoreFree":
			e.exportFilestore(metricName, metric, storageInfo.FileStoreSummary.FreeSpace, fileStoreType, fileStoreDir, ch)
		}
	}
}
