package collector

import (
	"strings"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

const (
	namespace = "artifactory"
)

var (
	defaultLabelNames     = []string{"node_id"}
	filestoreLabelNames   = append([]string{"storage_type", "storage_dir"}, defaultLabelNames...)
	repoLabelNames        = append([]string{"name", "type", "package_type"}, defaultLabelNames...)
	replicationLabelNames = append([]string{"name", "type", "url", "cron_exp", "status"}, defaultLabelNames...)
	federationLabelNames  = append([]string{"name", "remote_url", "remote_name"}, defaultLabelNames...)
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
		"users":  newMetric("users", "security", "Number of Artifactory users for each realm.", append([]string{"realm"}, defaultLabelNames...)),
		"groups": newMetric("groups", "security", "Number of Artifactory groups", defaultLabelNames),
	}

	storageMetrics = metrics{
		"artifacts":      newMetric("artifacts", "storage", "Total artifacts count stored in Artifactory.", defaultLabelNames),
		"artifactsSize":  newMetric("artifacts_size_bytes", "storage", "Total artifacts Size stored in Artifactory in bytes.", defaultLabelNames),
		"binaries":       newMetric("binaries", "storage", "Total binaries count stored in Artifactory.", defaultLabelNames),
		"binariesSize":   newMetric("binaries_size_bytes", "storage", "Total binaries Size stored in Artifactory in bytes.", defaultLabelNames),
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
		"healthy": newMetric("healthy", "system", "Is Artifactory working properly (1 = healthy).", defaultLabelNames),
		"version": newMetric("version", "system", "Version and revision of Artifactory as labels.", append([]string{"version", "revision"}, defaultLabelNames...)),
		"license": newMetric("license", "system", "License type and expiry as labels, seconds to expiration as value", append([]string{"type", "licensed_to", "expires"}, defaultLabelNames...)),
	}

	artifactsMetrics = metrics{
		"created1m":     newMetric("created_1m", "artifacts", "Number of artifacts created in the repository in the last 1 minute.", repoLabelNames),
		"created5m":     newMetric("created_5m", "artifacts", "Number of artifacts created in the repository in the last 5 minutes.", repoLabelNames),
		"created15m":    newMetric("created_15m", "artifacts", "Number of artifacts created in the repository in the last 15 minutes.", repoLabelNames),
		"downloaded1m":  newMetric("downloaded_1m", "artifacts", "Number of artifacts downloaded from the repository in the last 1 minute.", repoLabelNames),
		"downloaded5m":  newMetric("downloaded_5m", "artifacts", "Number of artifacts downloaded from the repository in the last 5 minutes.", repoLabelNames),
		"downloaded15m": newMetric("downloaded_15m", "artifacts", "Number of artifacts downloaded from the repository in the last 15 minutes.", repoLabelNames),
	}
	federationMetrics = metrics{
		"mirrorLag":         newMetric("mirror_lag", "federation", "Federation mirror lag in milliseconds.", federationLabelNames),
		"unavailableMirror": newMetric("unavailable_mirror", "federation", "Unsynchronized federated mirror status", append([]string{"status"}, federationLabelNames...)),
	}
	openMetrics = metrics{
		"openMetrics": newMetric("open_metrics", "openmetrics", "OpenMetrics proxied from Artifactory", append([]string{"metrics"}, defaultLabelNames...)),
	}
)

func init() {
	prometheus.MustRegister(version.NewCollector("artifactory_exporter"))
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
	if e.optionalMetrics.Artifacts {
		for _, m := range artifactsMetrics {
			ch <- m
		}
	}
	if e.optionalMetrics.FederationStatus {
		for _, m := range federationMetrics {
			ch <- m
		}
	}
	if e.optionalMetrics.OpenMetrics {
		for _, m := range openMetrics {
			ch <- m
		}
	}

	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.totalAPIErrors.Desc()
	ch <- e.jsonParseFailures.Desc()
}

// Collect fetches the stats from  Artifactory and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()

	up := e.scrape(ch)
	ch <- e.up
	e.up.Set(up)

	ch <- e.totalScrapes
	ch <- e.totalAPIErrors
	ch <- e.jsonParseFailures
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()

	// Collect License info
	var licenseType string
	license, err := e.client.FetchLicense()
	if err != nil {
		e.totalAPIErrors.Inc()
		return 0
	}
	licenseType = strings.ToLower(license.Type)
	// Some API endpoints are not available in OSS
	if licenseType != "oss" && licenseType != "jcr edition" && licenseType != "community edition for c/c++" {
		for metricName, metric := range securityMetrics {
			switch metricName {
			case "users":
				err := e.exportUsersCount(metricName, metric, ch)
				if err != nil {
					return 0
				}
			case "groups":
				err := e.exportGroups(metricName, metric, ch)
				if err != nil {
					return 0
				}
			}
		}
		err = e.exportReplications(ch)
		if err != nil {
			return 0
		}
	}

	// Collect and export open metrics
	if e.optionalMetrics.OpenMetrics {
		err = e.exportOpenMetrics(ch)
		if err != nil {
			return 0
		}
	}

	// Collect and export system metrics
	err = e.exportSystem(license, ch)
	if err != nil {
		return 0
	}

	// Fetch Storage Info stats and register them
	storageInfo, err := e.client.FetchStorageInfo()
	if err != nil {
		e.totalAPIErrors.Inc()
		return 0
	}
	e.exportStorage(storageInfo, ch)

	// Extract repo summaries from storageInfo and register them
	repoSummaryList, err := e.extractRepo(storageInfo)
	if err != nil {
		return 0
	}
	e.exportRepo(repoSummaryList, ch)

	// Get Downloaded and Created items for all repo in the last 1 and 5 minutes and add it to repoSummaryList
	if e.optionalMetrics.Artifacts {
		repoSummaryList, err = e.getTotalArtifacts(repoSummaryList)
		if err != nil {
			return 0
		}
		e.exportArtifacts(repoSummaryList, ch)
	}

	// Get Federation Mirror metrics
	if e.optionalMetrics.FederationStatus && e.client.IsFederationEnabled() {
		e.exportFederationMirrorLags(ch)
		e.exportFederationUnavailableMirrors(ch)
	}

	return 1
}
