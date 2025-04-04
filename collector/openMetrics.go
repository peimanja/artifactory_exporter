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

	openMetricsString := sanitizeOpenMetrics(openMetrics.PromMetrics)
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
