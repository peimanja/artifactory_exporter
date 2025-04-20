package collector

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
)

const (
	namespace = "artifactory"
)

// Label sets reused across metrics for consistency and reduced duplication
var (
	defaultLabelNames     = []string{"node_id"}
	filestoreLabelNames   = append([]string{"storage_type", "storage_dir"}, defaultLabelNames...)
	repoLabelNames        = append([]string{"name", "type", "package_type"}, defaultLabelNames...)
	replicationLabelNames = append([]string{"name", "type", "url", "cron_exp", "status"}, defaultLabelNames...)
	federationLabelNames  = append([]string{"name", "remote_url", "remote_name"}, defaultLabelNames...)
	certificateLabelNames = append([]string{"alias", "issued_by", "expires"}, defaultLabelNames...)
)

// Helper for creating new metric descriptors
func newMetric(metricName string, subsystem string, docString string, labelNames []string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, subsystem, metricName), docString, labelNames, nil)
}

type metrics map[string]*prometheus.Desc

// Metric descriptor groups by subsystem
var (
	replicationMetrics = metrics{
		"enabled": newMetric("enabled", "replication", "Replication status for an Artifactory repository (1 = enabled).", replicationLabelNames),
	}

	securityMetrics = metrics{
		"users":        newMetric("users", "security", "Number of Artifactory users for each realm.", append([]string{"realm"}, defaultLabelNames...)),
		"groups":       newMetric("groups", "security", "Number of Artifactory groups", defaultLabelNames),
		"certificates": newMetric("certificates", "security", "Internal SSL certificate information, seconds to expiration as value", certificateLabelNames),
	}

	storageMetrics = metrics{
		"artifacts":      newMetric("artifacts", "storage", "Total artifacts count stored in Artifactory.", defaultLabelNames),
		"artifactsSize":  newMetric("artifacts_size_bytes", "storage", "Total artifacts Size stored in Artifactory in bytes.", defaultLabelNames),
		"binaries":       newMetric("binaries", "storage", "Total binaries count stored in Artifactory.", defaultLabelNames),
		"binariesSize":   newMetric("binaries_size_bytes", "storage", "Total binaries Size stored in Artifactory in bytes.", defaultLabelNames),
		"filestore":      newMetric("filestore_bytes", "storage", "Total available space in the file store in bytes.", filestoreLabelNames),
		"filestoreUsed":  newMetric("filestore_used_bytes", "storage", "Used space in the file store in bytes.", filestoreLabelNames),
		"filestoreFree":  newMetric("filestore_free_bytes", "storage", "Free space in the file store in bytes.", filestoreLabelNames),
		"items":          newMetric("items", "storage", "Total items count stored in Artifactory.", defaultLabelNames),
		"repoUsed":       newMetric("repo_used_bytes", "storage", "Used space by an Artifactory repository in bytes.", repoLabelNames),
		"repoFolders":    newMetric("repo_folders", "storage", "Number of folders in an Artifactory repository.", repoLabelNames),
		"repoFiles":      newMetric("repo_files", "storage", "Number files in an Artifactory repository.", repoLabelNames),
		"repoItems":      newMetric("repo_items", "storage", "Number Items in an Artifactory repository.", repoLabelNames),
		"repoPercentage": newMetric("repo_percentage", "storage", "Percentage of space used by an Artifactory repository.", repoLabelNames),
	}

	systemMetrics = metrics{
		"healthy":  newMetric("healthy", "system", "Is Artifactory working properly (1 = healthy).", defaultLabelNames),
		"version":  newMetric("version", "system", "Version and revision of Artifactory as labels.", append([]string{"version", "revision"}, defaultLabelNames...)),
		"license":  newMetric("license", "system", "License type and expiry as labels, seconds to expiration as value", append([]string{"type", "licensed_to", "expires"}, defaultLabelNames...)),
		"licenses": newMetric("licenses", "system", "License type and expiry as labels, seconds to expiration as value", append([]string{"type", "valid_through", "licensed_to", "node_url", "license_hash", "expires"}, defaultLabelNames...)),
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
		"openMetrics": newMetric("open_metrics", "openmetrics", "OpenMetrics proxied from JFrog Platform", defaultLabelNames),
	}

	accessMetrics = metrics{
		"accessFederationValid": newMetric("access_federation_valid", "access", "Is JFrog Access Federation valid (1 = Circle of Trust validated)", defaultLabelNames),
	}
)

func init() {
	prometheus.MustRegister(version.NewCollector("artifactory_exporter"))
}

// Describe sends the descriptors of all metrics exported by the Artifactory exporter.
// Note: Metrics manually collected via Collect (like background task metrics)
// do not appear here, as they are registered and exported independently.
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
	if e.optionalMetrics.AccessFederationValidate {
		for _, m := range accessMetrics {
			ch <- m
		}
	}
	ch <- e.up.Desc()
	ch <- e.totalScrapes.Desc()
	ch <- e.totalAPIErrors.Desc()
	ch <- e.jsonParseFailures.Desc()
}

// Collect is called on each Prometheus scrape. It runs metric collection and publishes results.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.logger.Debug(">> Collect() fired")

	// Prevent concurrent scrapes from clashing with metric updates
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Execute data collection
	up := e.scrape(ch)

	// Export status and scrape counters
	ch <- e.up
	e.up.Set(up)
	ch <- e.totalScrapes
	ch <- e.totalAPIErrors
	ch <- e.jsonParseFailures

	// Manually collect background task metrics from the GaugeVec
	e.backgroundTaskMetrics.Collect(ch)
}

// scrape executes metric collection logic, split into helper functions to reduce complexity.
func (e *Exporter) scrape(ch chan<- prometheus.Metric) float64 {
	e.totalScrapes.Inc()

	// Reset the metric to avoid duplicate data
	e.backgroundTaskMetrics.Reset()

	if !e.runExportSteps(ch) {
		return 0
	}

	if e.optionalMetrics.BackgroundTasks {
		e.collectBackgroundTasks()
	}

	return 1
}

// runExportSteps performs the main metric collection sequence.
// Returns false if any required step fails.
func (e *Exporter) runExportSteps(ch chan<- prometheus.Metric) bool {
	if e.optionalMetrics.OpenMetrics {
		if err := e.exportOpenMetrics(ch); err != nil {
			return false
		}
	}
	if err := e.exportSystem(ch); err != nil {
		return false
	}
	if err := e.exportSystemHALicenses(ch); err != nil {
		return false
	}

	storageInfo, err := e.client.FetchStorageInfo()
	if err != nil {
		e.totalAPIErrors.Inc()
		return false
	}
	e.exportStorage(storageInfo, ch)

	repoSummaryList, err := e.extractRepo(storageInfo)
	if err != nil {
		return false
	}
	e.exportRepo(repoSummaryList, ch)

	if e.optionalMetrics.Artifacts {
		repoSummaryList, err = e.getTotalArtifacts(repoSummaryList)
		if err != nil {
			return false
		}
		e.exportArtifacts(repoSummaryList, ch)
	}

	if e.optionalMetrics.FederationStatus && e.client.IsFederationEnabled() {
		e.exportFederationMirrorLags(ch)
		e.exportFederationUnavailableMirrors(ch)
	}

	if e.optionalMetrics.AccessFederationValidate {
		e.exportAccessFederationValidate(ch)
	}

	return true
}

// collectBackgroundTasks emits a count of background tasks by (type, state) combination.
func (e *Exporter) collectBackgroundTasks() {
	e.logger.Debug("Collecting background tasks metrics")

	tasks, err := e.client.FetchBackgroundTasks()
	if err != nil {
		e.logger.Error("Error fetching background tasks", "error", err)
		return
	}

	// Collect background task metrics from Artifactory API
	// Use a map to count each (type, state) combo, and avoid duplicate label sets
	counter := make(map[[2]string]int)
	for _, task := range tasks {
		// Extract the class name only (e.g. "BundleCleanupJob") to reduce cardinality
		segments := strings.Split(task.Type, ".")
		shortType := segments[len(segments)-1]
		key := [2]string{shortType, task.State}
		counter[key]++
	}

	for key, count := range counter {
		e.backgroundTaskMetrics.WithLabelValues(key[0], key[1]).Set(float64(count))
	}
}
