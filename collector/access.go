package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) exportAccessFederationValidate(ch chan<- prometheus.Metric) error {
	// Fetch Federation Mirror Lags
	accessFederationValid, err := e.client.FetchAccessFederationValidStatus()
	if err != nil {
		e.logger.Warn(
			"JFrog Access Federation Circle of Trust was not successfully validated",
			"target", e.client.GetAccessFederationTarget(),
			"status", accessFederationValid.Status,
			"err", err.Error(),
		)
		e.totalAPIErrors.Inc()
	}
	value := convArtiToPromBool(accessFederationValid.Status)
	e.logger.Debug(
		logDbgMsgRegMetric,
		"metric", "accessFederationValid",
		"value", value,
	)
	ch <- prometheus.MustNewConstMetric(accessMetrics["accessFederationValid"], prometheus.GaugeValue, value, accessFederationValid.NodeId)
	return nil
}
