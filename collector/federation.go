package collector

import (
	"strings"

	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const FederationRepoType = "FEDERATED"

func (e *Exporter) exportFederationMirrorLags(ch chan<- prometheus.Metric) error {
	// Fetch Federation Mirror Lags
	federationMirrorLags, err := e.client.FetchMirrorLags()
	if err != nil {
		e.totalAPIErrors.Inc()
		return err
	}

	if len(federationMirrorLags.MirrorLags) == 0 {
		level.Debug(e.logger).Log("msg", "No federation mirror lags found")
		return nil
	}

	for _, mirrorLag := range federationMirrorLags.MirrorLags {
		level.Debug(e.logger).Log("msg", "Registering metric", "metric", "federationMirrorLag", "repo", mirrorLag.LocalRepoKey, "remote_url", mirrorLag.RemoteUrl, "remote_name", mirrorLag.RemoteRepoKey, "value", mirrorLag.LagInMS)
		ch <- prometheus.MustNewConstMetric(federationMetrics["mirrorLag"], prometheus.GaugeValue, float64(mirrorLag.LagInMS), mirrorLag.LocalRepoKey, mirrorLag.RemoteUrl, mirrorLag.RemoteRepoKey, federationMirrorLags.NodeId)
	}

	return nil
}

func (e *Exporter) exportFederationUnavailableMirrors(ch chan<- prometheus.Metric) error {
	// Fetch Federation Unavailable Mirrors
	federationUnavailableMirrors, err := e.client.FetchUnavailableMirrors()
	if err != nil {
		e.totalAPIErrors.Inc()
		return err
	}

	if len(federationUnavailableMirrors.UnavailableMirrors) == 0 {
		level.Debug(e.logger).Log("msg", "No federation unavailable mirrors found")
		return nil
	}

	for _, unavailableMirror := range federationUnavailableMirrors.UnavailableMirrors {
		level.Debug(e.logger).Log("msg", "Registering metric", "metric", "federationUnavailableMirror", "status", unavailableMirror.Status, "repo", unavailableMirror.LocalRepoKey, "remote_url", unavailableMirror.RemoteUrl, "remote_name", unavailableMirror.RemoteRepoKey)
		ch <- prometheus.MustNewConstMetric(federationMetrics["unavailableMirror"], prometheus.GaugeValue, 1, unavailableMirror.Status, unavailableMirror.LocalRepoKey, unavailableMirror.RemoteUrl, unavailableMirror.RemoteRepoKey, federationUnavailableMirrors.NodeId)
	}

	return nil
}

func (e *Exporter) getFederatedRepos(repoSummary []repoSummary) []string {
	var federatedRepos []string
	for _, repo := range repoSummary {
		if repo.Type == strings.ToLower(FederationRepoType) {
			federatedRepos = append(federatedRepos, repo.Name)
		}
	}
	return federatedRepos
}

func (e *Exporter) exportFederationRepoStatus(repoSummary []repoSummary, ch chan<- prometheus.Metric) error {
	repoList := e.getFederatedRepos(repoSummary)
	if len(repoList) == 0 {
		level.Debug(e.logger).Log("msg", "No federated repos found")
		return nil
	}

	for _, repo := range repoList {
		// Fetch Federation Repo Status
		federationRepoStatus, err := e.client.FetchFederatedRepoStatus(repo)
		if err != nil {
			e.totalAPIErrors.Inc()
			return err
		}

		// Check if the respond is not empty (depends on the Artifactory version)
		if federationRepoStatus.LocalKey == "" {
			level.Debug(e.logger).Log("msg", "No federation repo status found", "repo", repo)
			return nil
		}

		for _, mirrorEventsStatusInfo := range federationRepoStatus.MirrorEventsStatusInfo {
			level.Debug(e.logger).Log("msg", "Registering metric", "metric", "federationRepoStatus", "status", mirrorEventsStatusInfo.Status, "repo", repo, "remote_url", mirrorEventsStatusInfo.RemoteUrl, "remote_name", mirrorEventsStatusInfo.RemoteRepoKey)
			ch <- prometheus.MustNewConstMetric(federationMetrics["repoStatus"], prometheus.GaugeValue, 1, mirrorEventsStatusInfo.Status, repo, mirrorEventsStatusInfo.RemoteUrl, mirrorEventsStatusInfo.RemoteRepoKey, federationRepoStatus.NodeId)
		}
	}
	return nil
}
