package artifactory

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/peimanja/artifactory_exporter/config"
	l "github.com/peimanja/artifactory_exporter/logger"
)

func createFederationTestConfig() *config.Config {
	return &config.Config{
		ArtiScrapeURI:  "http://localhost:8081/artifactory",
		ArtiSSLVerify:  false,
		ArtiTimeout:    5 * time.Second,
		UseCache:       false,
		CacheTTL:       5 * time.Minute,
		CacheTimeout:   30 * time.Second,
		ListenAddress:  ":9531",
		MetricsPath:    "/metrics",
		Credentials:    &config.Credentials{AuthMethod: "userPass", Username: "user", Password: "pass"},
		Logger:         l.New(l.Config{Format: "logfmt", Level: "debug"}),
		ExporterRuntimeConfig: &config.ExporterRuntimeConfig{
			OptionalMetrics: config.OptionalMetrics{
				FederationStatus: true,
			},
		},
	}
}

func TestFetchUnavailableMirrors(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseCode   int
		expectedError  bool
		expectedMirrors int
		testDescription string
	}{
		{
			name:           "Normal JSON response",
			responseBody:   `{"unavailableMirrors":[{"repoKey":"test","status":"unavailable","localRepoKey":"local","remoteUrl":"http://remote","remoteRepoKey":"remote"}],"nodeId":"test-node"}`,
			responseCode:   200,
			expectedError:  false,
			expectedMirrors: 1,
			testDescription: "Should parse valid JSON response correctly",
		},
		{
			name:           "RTFS enabled response",
			responseBody:   "RTFS is enabled therefore get unavailable mirrors is not allowed",
			responseCode:   200,
			expectedError:  false,
			expectedMirrors: 0,
			testDescription: "Should handle RTFS enabled response gracefully without error",
		},
		{
			name:           "Empty JSON response",
			responseBody:   `{"unavailableMirrors":[],"nodeId":"test-node"}`,
			responseCode:   200,
			expectedError:  false,
			expectedMirrors: 0,
			testDescription: "Should handle empty mirrors list",
		},
		{
			name:           "Invalid JSON response",
			responseBody:   `{"invalid json`,
			responseCode:   200,
			expectedError:  true,
			expectedMirrors: 0,
			testDescription: "Should return error for malformed JSON (not RTFS message)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns the specified response
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Artifactory-Node-Id", "test-node")
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a client with the test server URL
			conf := createFederationTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			// Call the function
			result, err := client.FetchUnavailableMirrors()

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("%s: expected error but got none", tt.testDescription)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.testDescription, err)
			}

			// Check mirrors count
			if len(result.UnavailableMirrors) != tt.expectedMirrors {
				t.Errorf("%s: expected %d mirrors but got %d", tt.testDescription, tt.expectedMirrors, len(result.UnavailableMirrors))
			}
		})
	}
}

func TestFetchMirrorLags(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseCode   int
		expectedError  bool
		expectedLags   int
		testDescription string
	}{
		{
			name:           "Normal JSON response",
			responseBody:   `[{"localRepoKey":"local","remoteUrl":"http://remote","remoteRepoKey":"remote","lagInMS":100,"eventRegistrationTimeStamp":1234567890}]`,
			responseCode:   200,
			expectedError:  false,
			expectedLags:   1,
			testDescription: "Should parse valid JSON response correctly",
		},
		{
			name:           "RTFS enabled response",
			responseBody:   "RTFS is enabled therefore get mirror lags is not allowed",
			responseCode:   200,
			expectedError:  false,
			expectedLags:   0,
			testDescription: "Should handle RTFS enabled response gracefully without error",
		},
		{
			name:           "Empty JSON response",
			responseBody:   `[]`,
			responseCode:   200,
			expectedError:  false,
			expectedLags:   0,
			testDescription: "Should handle empty lags list",
		},
		{
			name:           "Invalid JSON response",
			responseBody:   `[{"invalid json`,
			responseCode:   200,
			expectedError:  true,
			expectedLags:   0,
			testDescription: "Should return error for malformed JSON (not RTFS message)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns the specified response
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Artifactory-Node-Id", "test-node")
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a client with the test server URL
			conf := createFederationTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			// Call the function
			result, err := client.FetchMirrorLags()

			// Check error expectation
			if tt.expectedError && err == nil {
				t.Errorf("%s: expected error but got none", tt.testDescription)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.testDescription, err)
			}

			// Check lags count
			if len(result.MirrorLags) != tt.expectedLags {
				t.Errorf("%s: expected %d lags but got %d", tt.testDescription, tt.expectedLags, len(result.MirrorLags))
			}
		})
	}
}

func TestIsFederationEnabled(t *testing.T) {
	tests := []struct {
		name         string
		responseBody string
		responseCode int
		expected     bool
		testDescription string
	}{
		{
			name:         "Federation enabled with JSON",
			responseBody: `{"unavailableMirrors":[],"nodeId":"test-node"}`,
			responseCode: 200,
			expected:     true,
			testDescription: "Should return true for successful JSON response",
		},
		{
			name:         "Federation enabled with RTFS",
			responseBody: "RTFS is enabled therefore get unavailable mirrors is not allowed",
			responseCode: 200,
			expected:     true,
			testDescription: "Should return true even when RTFS is enabled (federation is available but metrics are not)",
		},
		{
			name:         "Federation disabled (404)",
			responseBody: `{"errors":[{"status":404,"message":"Not Found"}]}`,
			responseCode: 404,
			expected:     false,
			testDescription: "Should return false for 404 (federation not available)",
		},
		{
			name:         "Server error",
			responseBody: `{"errors":[{"status":500,"message":"Internal Server Error"}]}`,
			responseCode: 500,
			expected:     false,
			testDescription: "Should return false for server errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that returns the specified response
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Artifactory-Node-Id", "test-node")
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a client with the test server URL
			conf := createFederationTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			// Call the function
			result := client.IsFederationEnabled()

			// Check result
			if result != tt.expected {
				t.Errorf("%s: expected %v but got %v", tt.testDescription, tt.expected, result)
			}
		})
	}
}