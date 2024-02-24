package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peimanja/artifactory_exporter/artifactory"
)

const msgErrCalcVal = "There was an issue calculating the value"

func (e *Exporter) exportCount(metricName string, metric *prometheus.Desc, count string, nodeId string, ch chan<- prometheus.Metric) {
	if count == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.convNumArtiToProm(count)
	if err != nil {
		e.jsonParseFailures.Inc()
		e.logger.Error(
			msgErrCalcVal,
			"metric", metricName,
			"err", err.Error(),
		)
		return
	}
	e.logger.Debug(
		logDbgMsgRegMetric,
		"metric", metricName,
		"value", value,
	)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value, nodeId)
}

func (e *Exporter) exportSize(metricName string, metric *prometheus.Desc, size string, nodeId string, ch chan<- prometheus.Metric) {
	if size == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, err := e.convNumArtiToProm(size)
	if err != nil {
		e.jsonParseFailures.Inc()
		e.logger.Error(
			msgErrCalcVal,
			"metric", metricName,
			"err", err.Error(),
		)
		return
	}
	e.logger.Debug(
		logDbgMsgRegMetric,
		"metric", metricName,
		"value", value,
	)
	ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, value, nodeId)
}

func (e *Exporter) exportFilestore(metricName string, metric *prometheus.Desc, size string, fileStoreType string, fileStoreDir string, nodeId string, ch chan<- prometheus.Metric) {
	if size == "" {
		e.jsonParseFailures.Inc()
		return
	}
	value, percent, err := e.convTwoNumsArtiToProm(size)
	/*
	 * What should you use the percentage for?
	 * Maybe Issue #126?
	 */
	if err != nil {
		e.jsonParseFailures.Inc()
		e.logger.Warn(
			msgErrCalcVal,
			"metric", metricName,
			"err", err.Error(),
		)
		return
	}
	e.logger.Debug(
		logDbgMsgRegMetric,
		"metric", metricName,
		"value", value,
		"percent", percent,
	)
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
	e.logger.Debug("Extracting repo summaries")
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
		rs.UsedSpace, err = e.convNumArtiToProm(repo.UsedSpace)
		if err != nil {
			e.logger.Warn(
				"There was an issue parsing repo UsedSpace",
				"repo", repo.RepoKey,
				"err", err.Error(),
			)
			e.jsonParseFailures.Inc()
			return repoSummaryList, err
		}
		if repo.Percentage == "N/A" {
			rs.Percentage = 0
		} else {
			/* WARNING!
			 * Previous e.removeCommas have been returning float from range [0.0, 100.0]
			 * Actual convNumArtiToProm returns float from range [0.0, 1.0]
			 * The application's behavior in this matter requires
			 * close observation in the near future.
			 */
			rs.Percentage, err = e.convNumArtiToProm(repo.Percentage)
			if err != nil {
				e.logger.Warn(
					"There was an issue parsing repo Percentage",
					"repo", repo.RepoKey,
					"err", err.Error(),
				)
				e.jsonParseFailures.Inc()
				return repoSummaryList, err
			}
		}
		rs.NodeId = storageInfo.NodeId
		repoSummaryList = append(repoSummaryList, rs)
	}
	return repoSummaryList, err
}

func (e *Exporter) exportRepo(repoSummaries []repoSummary, ch chan<- prometheus.Metric) {
	for _, repoSummary := range repoSummaries {
		for metricName, metric := range storageMetrics {
			switch metricName {
			case "repoUsed":
				e.logger.Debug(
					logDbgMsgRegMetric,
					"metric", metricName,
					"repo", repoSummary.Name,
					"type", repoSummary.Type,
					"package_type", repoSummary.PackageType,
					"value", repoSummary.UsedSpace,
				)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.UsedSpace, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoFolders":
				e.logger.Debug(
					logDbgMsgRegMetric,
					"metric", metricName,
					"repo", repoSummary.Name,
					"type", repoSummary.Type,
					"package_type", repoSummary.PackageType,
					"value", repoSummary.FoldersCount,
				)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.FoldersCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoItems":
				e.logger.Debug(
					logDbgMsgRegMetric,
					"metric", metricName,
					"repo", repoSummary.Name,
					"type", repoSummary.Type,
					"package_type", repoSummary.PackageType,
					"value", repoSummary.ItemsCount,
				)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.ItemsCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoFiles":
				e.logger.Debug(
					logDbgMsgRegMetric,
					"metric", metricName,
					"repo", repoSummary.Name,
					"type", repoSummary.Type,
					"package_type", repoSummary.PackageType,
					"value", repoSummary.FilesCount,
				)
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, repoSummary.FilesCount, repoSummary.Name, repoSummary.Type, repoSummary.PackageType, repoSummary.NodeId)
			case "repoPercentage":
				e.logger.Debug(
					logDbgMsgRegMetric,
					"metric", metricName,
					"repo", repoSummary.Name,
					"type", repoSummary.Type,
					"package_type", repoSummary.PackageType,
					"value", repoSummary.Percentage,
				)
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
