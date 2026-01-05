package collector

import (
	"fmt"
	"strings"

	"github.com/peimanja/artifactory_exporter/artifactory"
	"github.com/prometheus/client_golang/prometheus"
	ioPrometheusClient "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// sanitizeOpenMetrics sanitizes the OpenMetrics string
func sanitizeOpenMetrics(input string) string {
	lines := strings.Split(input, "\n")
	var result []string

	for _, line := range lines {
		if strings.HasPrefix(line, "# HELP") || strings.HasPrefix(line, "# TYPE") {
			line = strings.ReplaceAll(line, `\"`, `"`)
		}
		if strings.HasPrefix(line, "# EOF") {
			continue
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// processOpenMetrics is a modular helper function that processes OpenMetrics data
// and exports it to the prometheus metrics channel
func (e *Exporter) processOpenMetrics(openMetrics artifactory.OpenMetrics, ch chan<- prometheus.Metric, source string) error {
	openMetricsString := sanitizeOpenMetrics(openMetrics.PromMetrics)
	e.logger.Debug(
		fmt.Sprintf("OpenMetrics from %s", source),
		"body", openMetricsString,
	)

	parser := expfmt.TextParser{}
	metrics, err := parser.TextToMetricFamilies(strings.NewReader(openMetricsString))
	if err != nil {
		e.logger.Error(
			fmt.Sprintf("OpenMetrics downloaded from %s cannot be parsed using the \"github.com/prometheus/common/expfmt\".", source),
			"err", err.Error(),
			"response.body", openMetricsString,
		)
		return fmt.Errorf(
			"problem when parsing OpenMetrics downloaded from %s: %w",
			source, err,
		)
	}

	createDesc := func(fn, fh string, m *ioPrometheusClient.Metric) *prometheus.Desc {
		labels := make(map[string]string)
		for _, label := range m.Label {
			labels[*label.Name] = *label.Value
		}
		return prometheus.NewDesc(fn, fh, nil, labels)
	}

	for _, family := range metrics {
		fName := family.GetName()
		fHelp := family.GetHelp()
		if strings.HasPrefix(fName, "process_") { // avoid conflict with prometheus internal metrics
			fName = source + fName
		}
		for _, metric := range family.Metric {
			desc := createDesc(fName, fHelp, metric)
			switch family.GetType() {
			case ioPrometheusClient.MetricType_COUNTER:
				ch <- prometheus.MustNewConstMetric(
					desc,
					prometheus.CounterValue,
					metric.GetCounter().GetValue(),
				)
			case ioPrometheusClient.MetricType_GAUGE:
				ch <- prometheus.MustNewConstMetric(
					desc,
					prometheus.GaugeValue,
					metric.GetGauge().GetValue(),
				)
			}
		}
	}

	return nil
}
