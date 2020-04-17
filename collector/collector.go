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
	replicationLabelNames = []string{"name", "type", "url", "cron_exp"}
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
		"license": newMetric("license", "system", "License type and expiry as labels, seconds to expiration as value", []string{"type", "licensed_to", "expires"}),
	}

	artifactsMetrics = metrics{
		"created1m":     newMetric("created_1m", "artifacts", "Number of artifacts created in the repository in the last 1 minute.", repoLabelNames),
		"created5m":     newMetric("created_5m", "artifacts", "Number of artifacts created in the repository in the last 5 minutes.", repoLabelNames),
		"created15m":    newMetric("created_15m", "artifacts", "Number of artifacts created in the repository in the last 15 minutes.", repoLabelNames),
		"downloaded1m":  newMetric("downloaded_1m", "artifacts", "Number of artifacts downloaded from the repository in the last 1 minute.", repoLabelNames),
		"downloaded5m":  newMetric("downloaded_5m", "artifacts", "Number of artifacts downloaded from the repository in the last 5 minutes.", repoLabelNames),
		"downloaded15m": newMetric("downloaded_15m", "artifacts", "Number of artifacts downloaded from the repository in the last 15 minutes.", repoLabelNames),
	}

	artifactoryUp = newMetric("up", "", "Was the last scrape of Artifactory successful.", nil)
)

// Exporter collects JFrog Artifactory stats from the given URI and
// exports them using the prometheus metrics package.
type Exporter struct {
	URI        string
	cred       config.Credentials
	authMethod string
	sslVerify  bool
	timeout    time.Duration
	mutex      sync.RWMutex

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
	logger                          log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, cred config.Credentials, authMethod string, sslVerify bool, timeout time.Duration, logger log.Logger) (*Exporter, error) {

	return &Exporter{
		URI:        uri,
		cred:       cred,
		authMethod: authMethod,
		sslVerify:  sslVerify,
		timeout:    timeout,
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
	for _, m := range artifactsMetrics {
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

func (e *Exporter) fetchHTTP(uri string, path string, cred config.Credentials, authMethod string, sslVerify bool, timeout time.Duration) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/api/%s", uri, path)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	level.Debug(e.logger).Log("msg", "Fetching http", "path", fullPath)
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	if authMethod == "userPass" {
		req.SetBasicAuth(cred.Username, cred.Password)
	} else if authMethod == "accessToken" {
		req.Header.Add("Authorization", "Bearer "+cred.AccessToken)
	}
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
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching system/license", "err", err)
		return 0
	}
	licenseType = strings.ToLower(license.Type)
	healthy, err := e.fetchHealth()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching system/ping", "err", err)
		return 0
	}
	buildInfo, err := e.fetchBuildInfo()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching system/version", "err", err)
		return 0
	}

	for metricName, metric := range systemMetrics {
		switch metricName {
		case "healthy":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, healthy)
		case "version":
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, 1, buildInfo.Version, buildInfo.Revision)
		case "license":
			var validThrough float64
			timeNow := float64(time.Now().Unix())
			switch licenseType {
			case "oss":
				validThrough = timeNow
			default:
				if validThroughTime, err := time.Parse("Jan 2, 2006", license.ValidThrough); err != nil {
					level.Warn(e.logger).Log("msg", "Can't parse Artifactory license ValidThrough", "err", err)
					validThrough = timeNow
				} else {
					validThrough = float64(validThroughTime.Unix())
				}
			}
			ch <- prometheus.MustNewConstMetric(metric, prometheus.GaugeValue, validThrough-timeNow, licenseType, license.LicensedTo, license.ValidThrough)
		}
	}

	// Fetch Storage Info stats
	storageInfo, err := e.fetchStorageInfo()
	if err != nil {
		level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching storageinfo", "err", err)
		return 0
	}
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

	// Extract repo summaries from storageInfo and register them
	repoSummaryList, err := e.extractRepo(storageInfo)
	if err != nil {
		return 0
	}
	e.exportRepo(repoSummaryList, ch)

	// Get Downloaded and Created items for all repo in the last 1 and 5 minutes and add it to repoSummaryList
	repoSummaryList, err = e.getTotalArtifacts(repoSummaryList)
	if err != nil {
		return 0
	}
	e.exportArtifacts(repoSummaryList, ch)

	// Some API endpoints are not available in OSS
	if licenseType != "oss" {
		// Fetch Security stats
		users, err := e.fetchUsers()
		if err != nil {
			level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching security/users", "err", err)
			return 0
		}
		groups, err := e.fetchGroups()
		if err != nil {
			level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching security/groups", "err", err)
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
			level.Error(e.logger).Log("msg", "Can't scrape Artifactory when fetching replications", "err", err)
			return 0
		}

		e.exportReplications(replications, ch)
	}

	return 1
}
