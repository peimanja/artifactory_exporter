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

// createTestServer creates an HTTP test server with the given response
func createTestServer(responseBody string, responseCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Artifactory-Node-Id", "test-node")
		w.WriteHeader(responseCode)
		w.Write([]byte(responseBody))
	}))
}

// TestCase represents a common test case structure
type testCase struct {
	name            string
	responseBody    string
	responseCode    int
	expectedError   bool
	testDescription string
}

func TestFetchUnavailableMirrors(t *testing.T) {
	tests := []struct {
		testCase
		expectedMirrors int
	}{
		{
			testCase: testCase{
				name:            "Normal JSON response",
				responseBody:    `{"unavailableMirrors":[{"repoKey":"test","status":"unavailable","localRepoKey":"local","remoteUrl":"http://remote","remoteRepoKey":"remote"}],"nodeId":"test-node"}`,
				responseCode:    200,
				expectedError:   false,
				testDescription: "Should parse valid JSON response correctly",
			},
			expectedMirrors: 1,
		},
		{
			testCase: testCase{
				name:            "RTFS enabled response",
				responseBody:    "RTFS is enabled therefore get unavailable mirrors is not allowed",
				responseCode:    200,
				expectedError:   false,
				testDescription: "Should handle RTFS enabled response gracefully without error",
			},
			expectedMirrors: 0,
		},
		{
			testCase: testCase{
				name:            "Empty JSON response",
				responseBody:    `{"unavailableMirrors":[],"nodeId":"test-node"}`,
				responseCode:    200,
				expectedError:   false,
				testDescription: "Should handle empty mirrors list",
			},
			expectedMirrors: 0,
		},
		{
			testCase: testCase{
				name:            "Invalid JSON response",
				responseBody:    `{"invalid json`,
				responseCode:    200,
				expectedError:   true,
				testDescription: "Should return error for malformed JSON (not RTFS message)",
			},
			expectedMirrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(tt.responseBody, tt.responseCode)
			defer server.Close()

			conf := createFederationTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			result, err := client.FetchUnavailableMirrors()

			if tt.expectedError && err == nil {
				t.Errorf("%s: expected error but got none", tt.testDescription)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.testDescription, err)
			}

			if len(result.UnavailableMirrors) != tt.expectedMirrors {
				t.Errorf("%s: expected %d mirrors but got %d", tt.testDescription, tt.expectedMirrors, len(result.UnavailableMirrors))
			}
		})
	}
}

func TestFetchMirrorLags(t *testing.T) {
	tests := []struct {
		testCase
		expectedLags int
	}{
		{
			testCase: testCase{
				name:            "Normal JSON response",
				responseBody:    `[{"localRepoKey":"local","remoteUrl":"http://remote","remoteRepoKey":"remote","lagInMS":100,"eventRegistrationTimeStamp":1234567890}]`,
				responseCode:    200,
				expectedError:   false,
				testDescription: "Should parse valid JSON response correctly",
			},
			expectedLags: 1,
		},
		{
			testCase: testCase{
				name:            "RTFS enabled response",
				responseBody:    "RTFS is enabled therefore get mirror lags is not allowed",
				responseCode:    200,
				expectedError:   false,
				testDescription: "Should handle RTFS enabled response gracefully without error",
			},
			expectedLags: 0,
		},
		{
			testCase: testCase{
				name:            "Empty JSON response",
				responseBody:    `[]`,
				responseCode:    200,
				expectedError:   false,
				testDescription: "Should handle empty lags list",
			},
			expectedLags: 0,
		},
		{
			testCase: testCase{
				name:            "Invalid JSON response",
				responseBody:    `[{"invalid json`,
				responseCode:    200,
				expectedError:   true,
				testDescription: "Should return error for malformed JSON (not RTFS message)",
			},
			expectedLags: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(tt.responseBody, tt.responseCode)
			defer server.Close()

			conf := createFederationTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			result, err := client.FetchMirrorLags()

			if tt.expectedError && err == nil {
				t.Errorf("%s: expected error but got none", tt.testDescription)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("%s: expected no error but got: %v", tt.testDescription, err)
			}

			if len(result.MirrorLags) != tt.expectedLags {
				t.Errorf("%s: expected %d lags but got %d", tt.testDescription, tt.expectedLags, len(result.MirrorLags))
			}
		})
	}
}

func TestIsFederationEnabled(t *testing.T) {
	tests := []struct {
		testCase
		expected bool
	}{
		{
			testCase: testCase{
				name:            "Federation enabled with JSON",
				responseBody:    `{"unavailableMirrors":[],"nodeId":"test-node"}`,
				responseCode:    200,
				testDescription: "Should return true for successful JSON response",
			},
			expected: true,
		},
		{
			testCase: testCase{
				name:            "Federation enabled with RTFS",
				responseBody:    "RTFS is enabled therefore get unavailable mirrors is not allowed",
				responseCode:    200,
				testDescription: "Should return true even when RTFS is enabled (federation is available but metrics are not)",
			},
			expected: true,
		},
		{
			testCase: testCase{
				name:            "Federation disabled (404)",
				responseBody:    `{"errors":[{"status":404,"message":"Not Found"}]}`,
				responseCode:    404,
				testDescription: "Should return false for 404 (federation not available)",
			},
			expected: false,
		},
		{
			testCase: testCase{
				name:            "Server error",
				responseBody:    `{"errors":[{"status":500,"message":"Internal Server Error"}]}`,
				responseCode:    500,
				testDescription: "Should return false for server errors",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createTestServer(tt.responseBody, tt.responseCode)
			defer server.Close()

			conf := createFederationTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			result := client.IsFederationEnabled()

			if result != tt.expected {
				t.Errorf("%s: expected %v but got %v", tt.testDescription, tt.expected, result)
			}
		})
	}
}