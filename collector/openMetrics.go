package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

func (e *Exporter) exportOpenMetrics(ch chan<- prometheus.Metric) error {
	openMetrics, err := e.client.FetchOpenMetrics()
	if err != nil {
		e.logger.Error("There was an issue when try to fetch openMetrics")
		e.totalAPIErrors.Inc()
		return err
	}

	return e.processOpenMetrics(openMetrics, ch, "artifactory")
}

func (e *Exporter) exportXrayOpenMetrics(ch chan<- prometheus.Metric) error {
	openMetrics, err := e.client.FetchXrayOpenMetrics()
	if err != nil {
		e.logger.Error("There was an issue when try to fetch openMetrics")
		e.totalAPIErrors.Inc()
		return err
	}

	return e.processOpenMetrics(openMetrics, ch, "xray")
}
