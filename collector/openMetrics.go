package collector

import (
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	ioPrometheusClient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func (e *Exporter) exportOpenMetrics(ch chan<- prometheus.Metric) error {
	openMetrics, err := e.client.FetchOpenMetrics()
	if err != nil {
		e.logger.Error("There was an issue when try to fetch openMetrics")
		e.totalAPIErrors.Inc()
		return err
	}

	openMetricsString := openMetrics.PromMetrics
	e.logger.Debug(
		"OpenMetrics from Artifactory util",
		"body", openMetricsString,
	)
	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(strings.NewReader(openMetricsString))
	if err != nil {
		e.logger.Error(
			"Openmetrics downloaded from artifactory cannot be parsed using the “github.com/prometheus/common/expfmt”.",
			"err", err.Error(),
			"response.body", openMetricsString,
		)
		return fmt.Errorf(
			"problem when parsing openmetrics downloaded from artifactory: %w",
			err,
		)
	}

	for _, family := range metrics {
		for _, metric := range family.Metric {
			// create labels map
			labels := make(map[string]string)
			for _, label := range metric.Label {
				labels[*label.Name] = *label.Value
			}

			// create a new descriptor
			desc := prometheus.NewDesc(
				family.GetName(),
				family.GetHelp(),
				nil,
				labels,
			)

			// create a new metric and collect it
			switch family.GetType() {
			case ioPrometheusClient.MetricType_COUNTER:
				ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, metric.GetCounter().GetValue())
			case ioPrometheusClient.MetricType_GAUGE:
				ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, metric.GetGauge().GetValue())
			}
		}
	}

	return nil
}
