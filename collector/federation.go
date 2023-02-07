package collector

import (
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
