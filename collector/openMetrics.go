package collector

import (
	"strings"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func (e *Exporter) exportOpenMetrics(ch chan<- prometheus.Metric) error {
	// Fetch Open Metrics
	openMetrics, err := e.client.FetchOpenMetrics()
	if err != nil {
		level.Error(e.logger).Log("msg", "There was an issue when try to fetch openMetrics")
		e.totalAPIErrors.Inc()
		return err
	}

	level.Debug(e.logger).Log("msg", "OpenMetrics from Artifactory util", "body", openMetrics.PromMetrics)

	// assign openMetrics.Metric to a string variable
	openMetricsString := openMetrics.PromMetrics

	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(strings.NewReader(openMetricsString))
	if err != nil {
		// handle the error
		return err
	}

	for _, family := range metrics {
		for _, metric := range family.Metric {
			// create a new descriptor
			desc := prometheus.NewDesc(
				prometheus.BuildFQName("jfrog", "openmetrics", *family.Name),
				family.GetHelp(),
				nil, nil,
			)

			// create a new metric and collect it
			switch family.GetType() {
			case io_prometheus_client.MetricType_COUNTER:
				ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, metric.GetCounter().GetValue())
			case io_prometheus_client.MetricType_GAUGE:
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, metric.GetGauge().GetValue())
			}
		}
	}

	return nil
}
