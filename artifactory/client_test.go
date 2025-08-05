package artifactory

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/peimanja/artifactory_exporter/config"
	l "github.com/peimanja/artifactory_exporter/logger"
)

// Mock configuration for testing
func createTestConfig() *config.Config {
	return &config.Config{
		ArtiScrapeURI:   "http://localhost:8081/artifactory",
		ArtiSSLVerify:   false,
		ArtiTimeout:     5 * time.Second,
		UseCache:        false,
		CacheTTL:        5 * time.Minute,
		CacheTimeout:    30 * time.Second,
		OptionalMetrics: config.OptionalMetrics{},
		Credentials: &config.Credentials{
			AuthMethod: "userPass",
			Username:   "test",
			Password:   "test",
		},
		Logger: l.New(l.Config{Format: "logfmt", Level: "debug"}),
	}
}

func TestNewClient(t *testing.T) {
	conf := createTestConfig()
	client := NewClient(conf)

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.URI != conf.ArtiScrapeURI {
		t.Errorf("Client.URI = %s, want %s", client.URI, conf.ArtiScrapeURI)
	}

	if client.authMethod != conf.Credentials.AuthMethod {
		t.Errorf("Client.authMethod = %s, want %s", client.authMethod, conf.Credentials.AuthMethod)
	}

	if client.cred.Username != conf.Credentials.Username {
		t.Errorf("Client.cred.Username = %s, want %s", client.cred.Username, conf.Credentials.Username)
	}

	if client.client == nil {
		t.Error("Client.client should not be nil")
	}

	if client.logger == nil {
		t.Error("Client.logger should not be nil")
	}
}

func TestNewClientWithCache(t *testing.T) {
	conf := createTestConfig()
	conf.UseCache = true
	client := NewClient(conf)

	if client.responseCache == nil {
		t.Error("Client.responseCache should not be nil when UseCache is true")
	}
}

func TestGetAccessFederationTarget(t *testing.T) {
	conf := createTestConfig()
	conf.AccessFederationTarget = "https://example.com"
	client := NewClient(conf)

	target := client.GetAccessFederationTarget()
	if target != "https://example.com" {
		t.Errorf("GetAccessFederationTarget() = %s, want https://example.com", target)
	}
}

func TestFetchHTTPWithContext(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Artifactory-Node-Id", "test-node")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	}))
	defer server.Close()

	conf := createTestConfig()
	conf.ArtiScrapeURI = server.URL
	client := NewClient(conf)

	ctx := context.Background()
	resp, err := client.FetchHTTPWithContext(ctx, "system/ping")

	if err != nil {
		t.Fatalf("FetchHTTPWithContext() error = %v", err)
	}

	if resp == nil {
		t.Fatal("FetchHTTPWithContext() returned nil response")
	}

	if resp.NodeId != "test-node" {
		t.Errorf("Response.NodeId = %s, want test-node", resp.NodeId)
	}

	if string(resp.Body) != `{"status":"OK"}` {
		t.Errorf("Response.Body = %s, want {\"status\":\"OK\"}", string(resp.Body))
	}
}

func TestFetchHTTPWithContextTimeout(t *testing.T) {
	// Create a slow server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	conf := createTestConfig()
	conf.ArtiScrapeURI = server.URL
	client := NewClient(conf)

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.FetchHTTPWithContext(ctx, "system/ping")

	if err == nil {
		t.Error("Expected timeout error, but got none")
	}
}

func TestFetchBackgroundTasks(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse string
		statusCode     int
		expectedTasks  int
		expectError    bool
	}{
		{
			name: "Valid tasks response",
			serverResponse: `{
				"tasks": [
					{
						"id": "task1",
						"type": "CLEANUP",
						"state": "RUNNING",
						"description": "Cleanup task",
						"nodeId": "node1"
					},
					{
						"id": "task2",
						"type": "REPLICATION",
						"state": "SCHEDULED",
						"description": "Replication task",
						"nodeId": "node2"
					}
				]
			}`,
			statusCode:    http.StatusOK,
			expectedTasks: 2,
			expectError:   false,
		},
		{
			name:           "Empty tasks response",
			serverResponse: `{"tasks": []}`,
			statusCode:     http.StatusOK,
			expectedTasks:  0,
			expectError:    false,
		},
		{
			name:           "Invalid JSON response",
			serverResponse: `{"invalid": json}`,
			statusCode:     http.StatusOK,
			expectedTasks:  0,
			expectError:    true,
		},
		{
			name:           "Server error",
			serverResponse: `{"error": "Internal server error"}`,
			statusCode:     http.StatusInternalServerError,
			expectedTasks:  0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/tasks" {
					t.Errorf("Expected request to /api/tasks, got %s", r.URL.Path)
				}
				w.Header().Set("X-Artifactory-Node-Id", "test-node")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			conf := createTestConfig()
			conf.ArtiScrapeURI = server.URL
			client := NewClient(conf)

			tasks, err := client.FetchBackgroundTasks()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(tasks) != tt.expectedTasks {
				t.Errorf("Expected %d tasks, got %d", tt.expectedTasks, len(tasks))
			}

			if tt.expectedTasks > 0 {
				task := tasks[0]
				if task.ID == "" {
					t.Error("Task ID should not be empty")
				}
				if task.Type == "" {
					t.Error("Task Type should not be empty")
				}
				if task.State == "" {
					t.Error("Task State should not be empty")
				}
			}
		})
	}
}

func TestBackgroundTaskStructure(t *testing.T) {
	task := BackgroundTask{
		ID:          "test-id",
		Type:        "CLEANUP",
		State:       "RUNNING",
		Description: "Test cleanup task",
		NodeID:      "node-1",
	}

	// Test JSON marshaling
	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal BackgroundTask: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled BackgroundTask
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal BackgroundTask: %v", err)
	}

	if unmarshaled.ID != task.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, task.ID)
	}
	if unmarshaled.Type != task.Type {
		t.Errorf("Type mismatch: got %s, want %s", unmarshaled.Type, task.Type)
	}
	if unmarshaled.State != task.State {
		t.Errorf("State mismatch: got %s, want %s", unmarshaled.State, task.State)
	}
	if unmarshaled.Description != task.Description {
		t.Errorf("Description mismatch: got %s, want %s", unmarshaled.Description, task.Description)
	}
	if unmarshaled.NodeID != task.NodeID {
		t.Errorf("NodeID mismatch: got %s, want %s", unmarshaled.NodeID, task.NodeID)
	}
}

func TestClientConfiguration(t *testing.T) {
	t.Run("SSL verification enabled", func(t *testing.T) {
		conf := createTestConfig()
		conf.ArtiSSLVerify = true
		client := NewClient(conf)

		// We can't easily test the actual TLS config without complex setup,
		// but we can ensure the client was created successfully
		if client.client == nil {
			t.Error("Expected HTTP client to be created")
		}
	})

	t.Run("Custom timeout", func(t *testing.T) {
		conf := createTestConfig()
		conf.ArtiTimeout = 10 * time.Second
		client := NewClient(conf)

		if client.client.Timeout != 10*time.Second {
			t.Errorf("Client timeout = %v, want %v", client.client.Timeout, 10*time.Second)
		}
	})

	t.Run("Access token authentication", func(t *testing.T) {
		conf := createTestConfig()
		conf.Credentials.AuthMethod = "accessToken"
		conf.Credentials.AccessToken = "test-token"
		conf.Credentials.Username = ""
		conf.Credentials.Password = ""
		client := NewClient(conf)

		if client.authMethod != "accessToken" {
			t.Errorf("Expected auth method to be 'accessToken', got %s", client.authMethod)
		}
		if client.cred.AccessToken != "test-token" {
			t.Error("Access token not preserved in client")
		}
	})
}

func TestOptionalMetricsConfiguration(t *testing.T) {
	conf := createTestConfig()
	conf.OptionalMetrics = config.OptionalMetrics{
		Artifacts:                true,
		ReplicationStatus:        true,
		FederationStatus:         false,
		OpenMetrics:              true,
		AccessFederationValidate: false,
		BackgroundTasks:          true,
	}

	client := NewClient(conf)

	if !client.OptionalMetrics.Artifacts {
		t.Error("Artifacts metric should be enabled")
	}
	if !client.OptionalMetrics.ReplicationStatus {
		t.Error("ReplicationStatus metric should be enabled")
	}
	if client.OptionalMetrics.FederationStatus {
		t.Error("FederationStatus metric should be disabled")
	}
	if !client.OptionalMetrics.OpenMetrics {
		t.Error("OpenMetrics metric should be enabled")
	}
	if client.OptionalMetrics.AccessFederationValidate {
		t.Error("AccessFederationValidate metric should be disabled")
	}
	if !client.OptionalMetrics.BackgroundTasks {
		t.Error("BackgroundTasks metric should be enabled")
	}
}

// Test edge cases and error conditions
func TestClientEdgeCases(t *testing.T) {
	t.Run("Nil config handling", func(t *testing.T) {
		// This test ensures graceful handling of edge cases
		// Note: The actual NewClient function might panic with nil config,
		// which would be expected behavior
		defer func() {
			if r := recover(); r != nil {
				// Panic is acceptable for nil config
				t.Logf("NewClient panicked with nil config (expected): %v", r)
			}
		}()

		// This might panic, which is acceptable
		client := NewClient(nil)
		if client != nil {
			t.Log("NewClient handled nil config gracefully")
		}
	})

	t.Run("Empty URI handling", func(t *testing.T) {
		conf := createTestConfig()
		conf.ArtiScrapeURI = ""
		client := NewClient(conf)

		if client.URI != "" {
			t.Errorf("Expected empty URI to be preserved, got %s", client.URI)
		}
	})
}

// Test concurrent access safety
func TestClientConcurrency(t *testing.T) {
	conf := createTestConfig()
	client := NewClient(conf)

	// Test that multiple goroutines can safely access client methods
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			target := client.GetAccessFederationTarget()
			_ = target // Use the value to avoid unused variable
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestClientHTTPHeaders(t *testing.T) {
	headerChecks := make(map[string]string)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Store headers for verification
		for name, values := range r.Header {
			if len(values) > 0 {
				headerChecks[name] = values[0]
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	conf := createTestConfig()
	conf.ArtiScrapeURI = server.URL
	client := NewClient(conf)

	// Make a request to trigger header inspection
	ctx := context.Background()
	_, err := client.FetchHTTPWithContext(ctx, "test")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Check that User-Agent header is set
	if userAgent, exists := headerChecks["User-Agent"]; !exists {
		t.Error("User-Agent header should be set")
	} else if userAgent == "" {
		t.Error("User-Agent header should not be empty")
	}
}
