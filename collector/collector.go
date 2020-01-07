package collector

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "artifactory"
)

var (
	filestoreLabelNames   = []string{"storage_type", "storage_dir"}
	repoLabelNames        = []string{"name", "type", "package_type"}
	replicationLabelNames = []string{"name", "type", "cron_exp"}
)

func newMetric(metricName string, subsystem string, docString string, labelNames []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, metricName), docString, labelNames, nil)
}

type metrics map[string]*prometheus.Desc

var (
	replicationMetrics = metrics{
		"enabled": newMetric("enabled", "replication", "Replication status for an Artifactory repository (1 = enabled).", replicationLabelNames),
	}

	securityMetrics = metrics{
		"users":  newMetric("users", "security", "Number of Artifactory users for each realm.", []string{"realm"}),
		"groups": newMetric("groups", "security", "Number of Artifactory groups", nil),
	}

	storageMetrics = metrics{
		"artifacts":      newMetric("artifacts", "storage", "Total artifacts count stored in Artifactory.", nil),
		"artifactsSize":  newMetric("artifacts_size_bytes", "storage", "Total artifacts Size stored in Artifactory in bytes.", nil),
		"binaries":       newMetric("binaries", "storage", "Total binaries count stored in Artifactory.", nil),
		"binariesSize":   newMetric("binaries_size_bytes", "storage", "Total binaries Size stored in Artifactory in bytes.", nil),
		"filestore":      newMetric("filestore_bytes", "storage", "Total available space in the file store in bytes.", filestoreLabelNames),
		"filestoreUsed":  newMetric("filestore_used_bytes", "storage", "Used space in the file store in bytes.", filestoreLabelNames),
		"filestoreFree":  newMetric("filestore_free_bytes", "storage", "Free space in the file store in bytes.", filestoreLabelNames),
		"repoUsed":       newMetric("repo_used_bytes", "storage", "Used space by an Artifactory repository in bytes.", repoLabelNames),
		"repoFolders":    newMetric("repo_folders", "storage", "Number of folders in an Artifactory repository.", repoLabelNames),
		"repoFiles":      newMetric("repo_files", "storage", "Number files in an Artifactory repository.", repoLabelNames),
		"repoItems":      newMetric("repo_items", "storage", "Number Items in an Artifactory repository.", repoLabelNames),
		"repoPercentage": newMetric("repo_percentage", "storage", "Percentage of space used by an Artifactory repository.", repoLabelNames),
	}

	systemMetrics = metrics{
		"healthy": newMetric("healthy", "system", "Is Artifactory working properly (1 = healthy).", nil),
		"version": newMetric("version", "system", "Version and revision of Artifactory as labels.", []string{"version", "revision"}),
		"license": newMetric("license", "system", "License type and expiry as labels", []string{"type", "licensed_to", "expires"}),
	}

	artifactoryUp = newMetric("up", "", "Was the last scrape of Artifactory successful.", nil)
)

// Exporter collects JFrog Artifactory stats from the given URI and
// exports them using the prometheus metrics package.
type Exporter struct {
	URI       string
	bc        config.BasicCredentials
	sslVerify bool
	timeout   time.Duration
	mutex     sync.RWMutex

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
	logger                          log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, bc config.BasicCredentials, sslVerify bool, timeout time.Duration, logger log.Logger) (*Exporter, error) {

	return &Exporter{
		URI:       uri,
		bc:        bc,
		sslVerify: sslVerify,
		timeout:   timeout,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "Was the last scrape of artifactory successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_total_scrapes",
			Help:      "Current total artifactory scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "exporter_json_parse_failures",
			Help:      "Number of errors while parsing Json.",
		}),
		logger: logger,
	}, nil
}

// Describe describes all the metrics ever exported by the Artifactory exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range replicationMetrics {
		ch <- m
	}
	for _, m := range securityMetrics {
		ch <- m
	}
	for _, m := range storageMetrics {
		ch <- m
	}
	for _, m := range systemMetrics {
		ch <- m
	}
	ch <- artifactoryUp
	ch <- e.totalScrapes.Desc()
	ch <- e.jsonParseFailures.Desc()
}

// Collect fetches the stats from  Artifactiry and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(artifactoryUp, prometheus.GaugeValue, up)
	ch <- e.totalScrapes
	ch <- e.jsonParseFailures
}

func fetchHTTP(uri string, path string, bc config.BasicCredentials, sslVerify bool, timeout time.Duration) ([]byte, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", uri+"/api/"+path, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(bc.Username, bc.Password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil

}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()

	// Fetch System stats
	var licenseType string
	license, err := e.fetchLicense()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
		return 0
	}
	licenseType = strings.ToLower(license.Type)
	healthy, err := e.fetchHealth()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
		return 0
	}
	buildInfo, err := e.fetchBuildInfo()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
		return 0
	}

	for metricName, metric := range systemMetrics {
		switch metricName {
		case "healthy":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, healthy)
		case "version":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, buildInfo.Version, buildInfo.Revision)
		case "license":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, licenseType, license.LicensedTo, license.ValidThrough)
		}
	}

	// Fetch Storage Info stats
	storageInfo, err := e.fetchStorageInfo()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
		return 0
	}
	fileStoreType := strings.ToLower(storageInfo.StorageSummary.FileStoreSummary.StorageType)
	fileStoreDir := storageInfo.StorageSummary.FileStoreSummary.StorageDirectory
	for metricName, metric := range storageMetrics {
		switch metricName {
		case "artifacts":
			e.exportCount(metricName, metric, storageInfo.StorageSummary.BinariesSummary.ArtifactsCount, ch)
		case "artifactsSize":
			e.exportSize(metricName, metric, storageInfo.StorageSummary.BinariesSummary.ArtifactsSize, ch)
		case "binaries":
			e.exportCount(metricName, metric, storageInfo.StorageSummary.BinariesSummary.BinariesCount, ch)
		case "binariesSize":
			e.exportSize(metricName, metric, storageInfo.StorageSummary.BinariesSummary.BinariesSize, ch)
		case "filestore":
			e.exportFilestore(metricName, metric, storageInfo.StorageSummary.FileStoreSummary.TotalSpace, fileStoreType, fileStoreDir, ch)
		case "filestoreUsed":
			e.exportFilestore(metricName, metric, storageInfo.StorageSummary.FileStoreSummary.UsedSpace, fileStoreType, fileStoreDir, ch)
		case "filestoreFree":
			e.exportFilestore(metricName, metric, storageInfo.StorageSummary.FileStoreSummary.FreeSpace, fileStoreType, fileStoreDir, ch)
		}
	}
	e.extractRepoSummary(storageInfo, ch)

	// Some API endpoints are not available in OSS
	if licenseType != "oss" {
		// Fetch Security stats
		users, err := e.fetchUsers()
		if err != nil {
			level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
			return 0
		}
		groups, err := e.fetchGroups()
		if err != nil {
			level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
			return 0
		}

		for metricName, metric := range securityMetrics {
			switch metricName {
			case "users":
				e.countUsers(metricName, metric, users, ch)
			case "groups":
				ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, float64(len(groups)))
			}
		}

		// Fetch Replications stats
		replications, err := e.fetchReplications()
		if err != nil {
			level.Error(e.logger).Log("msg", "Can't scrape Artifactory", "err", err)
			return 0
		}

		e.exportReplications(replications, ch)
	}

	return 1
}
