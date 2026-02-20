package collector

import (
    "fmt"
    "sort"
    "strings"

    "github.com/prometheus/client_golang/prometheus"
    ioPrometheusClient "github.com/prometheus/client_model/go"
    "github.com/prometheus/common/expfmt"
)

func (e *Exporter) exportOpenMetrics(ch chan<- prometheus.Metric) error {
    metrics, err := e.fetchAndParseMetrics()
    if err != nil {
        return err
    }

    sentMetrics := make(map[string]bool)
    
    for _, family := range metrics {
        e.processMetricFamily(family, ch, sentMetrics)
    }

    return nil
}

func (e *Exporter) fetchAndParseMetrics() (map[string]*ioPrometheusClient.MetricFamily, error) {
    openMetrics, err := e.client.FetchOpenMetrics()
    if err != nil {
        e.logger.Error("There was an issue when try to fetch openMetrics")
        e.totalAPIErrors.Inc()
        return nil, err
    }

    openMetricsString := sanitizeOpenMetrics(openMetrics.PromMetrics)
    e.logger.Debug("OpenMetrics from Artifactory util", "body", openMetricsString)
    
    parser := expfmt.TextParser{}
    metrics, err := parser.TextToMetricFamilies(strings.NewReader(openMetricsString))
    if err != nil {
        e.logger.Error(
            "Openmetrics downloaded from artifactory cannot be parsed using the expfmt parser",
            "err", err.Error(),
            "response.body", openMetricsString,
        )
        return nil, fmt.Errorf("problem when parsing openmetrics downloaded from artifactory: %w", err)
    }

    return metrics, nil
}

func (e *Exporter) processMetricFamily(family *ioPrometheusClient.MetricFamily, ch chan<- prometheus.Metric, sentMetrics map[string]bool) {
    fName := family.GetName()
    fHelp := fixHelpText(fName, family.GetHelp())
    
    for _, metric := range family.Metric {
        e.processMetric(fName, fHelp, family.GetType(), metric, ch, sentMetrics)
    }
}

func (e *Exporter) processMetric(fName, fHelp string, metricType ioPrometheusClient.MetricType, metric *ioPrometheusClient.Metric, ch chan<- prometheus.Metric, sentMetrics map[string]bool) {
    metricKey := createMetricKey(fName, metric)
    if sentMetrics[metricKey] {
        return
    }

    desc := createMetricDesc(fName, fHelp, metric)
    promMetric := convertToPrometheusMetric(desc, metricType, metric)
    
    if promMetric != nil {
        ch <- promMetric
        sentMetrics[metricKey] = true
    }
}

func fixHelpText(name, help string) string {
    if name == "process_start_time_seconds" && help == "Start time of the process since unix epoch." {
        return "Start time of the process since unix epoch in seconds."
    }
    return help
}

func createMetricDesc(name, help string, m *ioPrometheusClient.Metric) *prometheus.Desc {
    labels := make(map[string]string)
    for _, label := range m.Label {
        labels[*label.Name] = *label.Value
    }
    return prometheus.NewDesc(name, help, nil, labels)
}

func convertToPrometheusMetric(desc *prometheus.Desc, metricType ioPrometheusClient.MetricType, metric *ioPrometheusClient.Metric) prometheus.Metric {
    switch metricType {
    case ioPrometheusClient.MetricType_COUNTER:
        return prometheus.MustNewConstMetric(desc, prometheus.CounterValue, metric.GetCounter().GetValue())
    case ioPrometheusClient.MetricType_GAUGE:
        return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, metric.GetGauge().GetValue())
    default:
        return nil
    }
}

// createMetricKey creates a unique key for a metric based on name and labels
func createMetricKey(name string, metric *ioPrometheusClient.Metric) string {
    var labels []string
    for _, label := range metric.Label {
        labels = append(labels, *label.Name+"="+*label.Value)
    }
    sort.Strings(labels)
    return name + "{" + strings.Join(labels, ",") + "}"
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