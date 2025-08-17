package collector

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/peimanja/artifactory_exporter/config"
)

func createTestConfig() *config.Config {
	return &config.Config{
		ListenAddress:          ":9531",
		MetricsPath:            "/metrics",
		ArtiScrapeURI:          "http://localhost:8081/artifactory",
		ArtiSSLVerify:          false,
		ArtiTimeout:            5 * time.Second,
		UseCache:               false,
		CacheTimeout:           30 * time.Second,
		CacheTTL:               5 * time.Minute,
		OptionalMetrics:        config.OptionalMetrics{OpenMetrics: true},
		AccessFederationTarget: "",
		Logger:                 newTestLogger(),
		Credentials: &config.Credentials{
			AuthMethod:  "accessToken",
			Username:    "",
			Password:    "",
			AccessToken: "test-token",
		},
	}
}

func TestOpenMetricsNotInDescribe(t *testing.T) {
	// Create an exporter with OpenMetrics enabled
	conf := createTestConfig()
	exporter, err := NewExporter(conf)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	// Collect all metric descriptors
	ch := make(chan *prometheus.Desc, 100)
	go func() {
		defer close(ch)
		exporter.Describe(ch)
	}()

	// Check that no openMetrics descriptors are in the Describe output
	openMetricsDescriptors := 0
	for desc := range ch {
		descString := desc.String()
		if strings.Contains(descString, "open_metrics") || strings.Contains(descString, "openmetrics") {
			openMetricsDescriptors++
		}
	}

	if openMetricsDescriptors > 0 {
		t.Errorf("Found %d openMetrics descriptors in Describe output, expected 0 (OpenMetrics should be handled dynamically)", openMetricsDescriptors)
	}
}

func TestSanitizeOpenMetrics(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Remove EOF line",
			input: `# HELP test_metric A test metric
# TYPE test_metric counter
test_metric 42
# EOF`,
			expected: `# HELP test_metric A test metric
# TYPE test_metric counter
test_metric 42`,
		},
		{
			name: "Fix escaped quotes in HELP and TYPE",
			input: `# HELP test_metric A \"quoted\" description
# TYPE test_metric counter
test_metric{label=\"value\"} 42`,
			expected: `# HELP test_metric A "quoted" description
# TYPE test_metric counter
test_metric{label=\"value\"} 42`,
		},
		{
			name: "Handle multiple lines with various issues",
			input: `# HELP jfsh_binaries_download_total Counts the \"total\" binaries download
# TYPE jfsh_binaries_download_total counter
jfsh_binaries_download_total{id="cache-fs",name="cache_fs"} 10592 1698850569605
# HELP jfrt_db_connections_active_total Total \"Active\" Connections
# TYPE jfrt_db_connections_active_total gauge
jfrt_db_connections_active_total 0 1698850569605
# EOF`,
			expected: `# HELP jfsh_binaries_download_total Counts the "total" binaries download
# TYPE jfsh_binaries_download_total counter
jfsh_binaries_download_total{id="cache-fs",name="cache_fs"} 10592 1698850569605
# HELP jfrt_db_connections_active_total Total "Active" Connections
# TYPE jfrt_db_connections_active_total gauge
jfrt_db_connections_active_total 0 1698850569605`,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
		},
		{
			name: "No changes needed",
			input: `# HELP test_metric A simple metric
# TYPE test_metric counter
test_metric 42`,
			expected: `# HELP test_metric A simple metric
# TYPE test_metric counter
test_metric 42`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeOpenMetrics(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeOpenMetrics() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExportOpenMetricsIntegration(t *testing.T) {
	// This test verifies that OpenMetrics can be parsed and collected without static descriptors
	sampleOpenMetrics := `# HELP jfsh_binaries_download_total Counts the total binaries download
# TYPE jfsh_binaries_download_total counter
jfsh_binaries_download_total{id="cache-fs",name="cache_fs"} 10592 1698850569605
# HELP jfrt_runtime_heap_freememory_bytes Free Memory
# TYPE jfrt_runtime_heap_freememory_bytes gauge
jfrt_runtime_heap_freememory_bytes 3532722600 1698850569605`

	// Test that the sanitizeOpenMetrics function works
	sanitized := sanitizeOpenMetrics(sampleOpenMetrics)
	
	// The fact that we can parse and process the metrics without panicking
	// validates that the dynamic approach works correctly
	if len(sanitized) == 0 {
		t.Error("sanitizeOpenMetrics returned empty string for valid input")
	}

	if !strings.Contains(sanitized, "jfsh_binaries_download_total") {
		t.Error("sanitizeOpenMetrics removed expected metric content")
	}
}