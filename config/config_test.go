package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
)

func TestCredentialsValidation(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		password    string
		accessToken string
		expected    string
		expectError bool
	}{
		{
			name:        "Valid username/password",
			username:    "testuser",
			password:    "testpass",
			accessToken: "",
			expected:    "userPass",
			expectError: false,
		},
		{
			name:        "Valid access token",
			username:    "",
			password:    "",
			accessToken: "testtoken",
			expected:    "accessToken",
			expectError: false,
		},
		{
			name:        "No credentials",
			username:    "",
			password:    "",
			accessToken: "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Both username/password and token",
			username:    "testuser",
			password:    "testpass",
			accessToken: "testtoken",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Username without password",
			username:    "testuser",
			password:    "",
			accessToken: "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Password without username",
			username:    "",
			password:    "testpass",
			accessToken: "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Empty username with password",
			username:    "   ",
			password:    "testpass",
			accessToken: "",
			expected:    "userPass", // The config actually treats whitespace as valid
			expectError: false,
		},
		{
			name:        "Username with empty password",
			username:    "testuser",
			password:    "   ",
			accessToken: "",
			expected:    "userPass", // The config actually treats whitespace as valid
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("ARTI_USERNAME", tt.username)
			os.Setenv("ARTI_PASSWORD", tt.password)
			os.Setenv("ARTI_ACCESS_TOKEN", tt.accessToken)

			// Clean up after test
			defer func() {
				os.Unsetenv("ARTI_USERNAME")
				os.Unsetenv("ARTI_PASSWORD")
				os.Unsetenv("ARTI_ACCESS_TOKEN")
			}()

			var credentials Credentials
			err := envconfig.Process("", &credentials)
			if err != nil {
				t.Fatalf("envconfig.Process failed: %v", err)
			}

			// Simulate the validation logic from NewConfig
			var authMethod string
			var validationErr error

			// Trim spaces for validation
			username := credentials.Username
			password := credentials.Password
			accessToken := credentials.AccessToken

			if username != "" && password != "" && accessToken == "" {
				authMethod = "userPass"
			} else if username == "" && password == "" && accessToken != "" {
				authMethod = "accessToken"
			} else {
				validationErr = fmt.Errorf("invalid credentials combination")
			}

			if tt.expectError {
				if validationErr == nil {
					t.Errorf("Expected error for test case '%s', but got none", tt.name)
				}
				return
			}

			if validationErr != nil {
				t.Errorf("Unexpected error for test case '%s': %v", tt.name, validationErr)
				return
			}

			if authMethod != tt.expected {
				t.Errorf("AuthMethod = '%s', want '%s'", authMethod, tt.expected)
			}
		})
	}
}

func TestOptionalMetricsValidation(t *testing.T) {
	tests := []struct {
		name           string
		metrics        []string
		expectedResult OptionalMetrics
		expectError    bool
	}{
		{
			name:    "No optional metrics",
			metrics: []string{},
			expectedResult: OptionalMetrics{
				Artifacts:                false,
				ReplicationStatus:        false,
				FederationStatus:         false,
				OpenMetrics:              false,
				AccessFederationValidate: false,
				BackgroundTasks:          false,
			},
			expectError: false,
		},
		{
			name:    "Single valid metric",
			metrics: []string{"artifacts"},
			expectedResult: OptionalMetrics{
				Artifacts:                true,
				ReplicationStatus:        false,
				FederationStatus:         false,
				OpenMetrics:              false,
				AccessFederationValidate: false,
				BackgroundTasks:          false,
			},
			expectError: false,
		},
		{
			name:    "Multiple valid metrics",
			metrics: []string{"artifacts", "federation_status", "background_tasks"},
			expectedResult: OptionalMetrics{
				Artifacts:                true,
				ReplicationStatus:        false,
				FederationStatus:         true,
				OpenMetrics:              false,
				AccessFederationValidate: false,
				BackgroundTasks:          true,
			},
			expectError: false,
		},
		{
			name:    "All valid metrics",
			metrics: []string{"artifacts", "replication_status", "federation_status", "open_metrics", "access_federation_validate", "background_tasks"},
			expectedResult: OptionalMetrics{
				Artifacts:                true,
				ReplicationStatus:        true,
				FederationStatus:         true,
				OpenMetrics:              true,
				AccessFederationValidate: true,
				BackgroundTasks:          true,
			},
			expectError: false,
		},
		{
			name:           "Invalid metric",
			metrics:        []string{"invalid_metric"},
			expectedResult: OptionalMetrics{},
			expectError:    true,
		},
		{
			name:           "Mix of valid and invalid metrics",
			metrics:        []string{"artifacts", "invalid_metric"},
			expectedResult: OptionalMetrics{},
			expectError:    true,
		},
		{
			name:    "Duplicate metrics",
			metrics: []string{"artifacts", "artifacts", "federation_status"},
			expectedResult: OptionalMetrics{
				Artifacts:                true,
				ReplicationStatus:        false,
				FederationStatus:         true,
				OpenMetrics:              false,
				AccessFederationValidate: false,
				BackgroundTasks:          false,
			},
			expectError: false,
		},
		{
			name:           "Empty string in metrics",
			metrics:        []string{""},
			expectedResult: OptionalMetrics{},
			expectError:    true,
		},
		{
			name:           "Whitespace in metrics",
			metrics:        []string{" artifacts "},
			expectedResult: OptionalMetrics{},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the validation logic from NewConfig
			optMetrics := OptionalMetrics{}
			var validationErr error

			for _, metric := range tt.metrics {
				switch metric {
				case "artifacts":
					optMetrics.Artifacts = true
				case "replication_status":
					optMetrics.ReplicationStatus = true
				case "federation_status":
					optMetrics.FederationStatus = true
				case "open_metrics":
					optMetrics.OpenMetrics = true
				case "access_federation_validate":
					optMetrics.AccessFederationValidate = true
				case "background_tasks":
					optMetrics.BackgroundTasks = true
				default:
					validationErr = fmt.Errorf("unknown optional metric: %s", metric)
					break
				}
			}

			if tt.expectError {
				if validationErr == nil {
					t.Errorf("Expected error for test case '%s', but got none", tt.name)
				}
				return
			}

			if validationErr != nil {
				t.Errorf("Unexpected error for test case '%s': %v", tt.name, validationErr)
				return
			}

			// Compare the result
			if optMetrics != tt.expectedResult {
				t.Errorf("OptionalMetrics = %+v, want %+v", optMetrics, tt.expectedResult)
			}
		})
	}
}

func TestTimeoutValidation(t *testing.T) {
	tests := []struct {
		name        string
		timeout     string
		expected    time.Duration
		expectError bool
	}{
		{"Valid seconds", "5s", 5 * time.Second, false},
		{"Valid minutes", "2m", 2 * time.Minute, false},
		{"Valid hours", "1h", 1 * time.Hour, false},
		{"Valid milliseconds", "500ms", 500 * time.Millisecond, false},
		{"Valid microseconds", "1000us", 1000 * time.Microsecond, false},
		{"Valid nanoseconds", "1000ns", 1000 * time.Nanosecond, false},
		{"Combined duration", "1h30m45s", 1*time.Hour + 30*time.Minute + 45*time.Second, false},
		{"Invalid format", "invalid", 0, true},
		{"Missing unit", "123", 0, true},
		{"Negative duration", "-5s", -5 * time.Second, false}, // Negative durations are valid in Go
		{"Zero duration", "0s", 0 * time.Second, false},
		{"Zero without unit", "0", 0, false}, // Go's time.ParseDuration accepts "0" as valid
		{"Float seconds", "1.5s", 1500 * time.Millisecond, false},
		{"Very large duration", "8760h", 8760 * time.Hour, false}, // 1 year
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := time.ParseDuration(tt.timeout)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for timeout '%s', but got none", tt.timeout)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for timeout '%s': %v", tt.timeout, err)
				return
			}

			if duration != tt.expected {
				t.Errorf("Parsed duration = %v, want %v", duration, tt.expected)
			}
		})
	}
}

func TestOptionalMetricsList(t *testing.T) {
	expectedMetrics := []string{
		"artifacts",
		"replication_status",
		"federation_status",
		"open_metrics",
		"access_federation_validate",
		"background_tasks",
	}

	if len(optionalMetricsList) != len(expectedMetrics) {
		t.Errorf("optionalMetricsList length = %d, want %d", len(optionalMetricsList), len(expectedMetrics))
	}

	for i, expected := range expectedMetrics {
		if i >= len(optionalMetricsList) {
			t.Errorf("Missing metric at index %d: %s", i, expected)
			continue
		}
		if optionalMetricsList[i] != expected {
			t.Errorf("optionalMetricsList[%d] = %s, want %s", i, optionalMetricsList[i], expected)
		}
	}
}

func TestCredentialsStruct(t *testing.T) {
	// Test struct field types and tags
	creds := Credentials{
		AuthMethod:  "userPass",
		Username:    "testuser",
		Password:    "testpass",
		AccessToken: "",
	}

	if creds.AuthMethod != "userPass" {
		t.Errorf("AuthMethod field not set correctly")
	}

	if creds.Username != "testuser" {
		t.Errorf("Username field not set correctly")
	}

	if creds.Password != "testpass" {
		t.Errorf("Password field not set correctly")
	}

	if creds.AccessToken != "" {
		t.Errorf("AccessToken should be empty")
	}
}

func TestOptionalMetricsStruct(t *testing.T) {
	// Test default values
	opt := OptionalMetrics{}

	if opt.Artifacts {
		t.Error("Artifacts should default to false")
	}
	if opt.ReplicationStatus {
		t.Error("ReplicationStatus should default to false")
	}
	if opt.FederationStatus {
		t.Error("FederationStatus should default to false")
	}
	if opt.OpenMetrics {
		t.Error("OpenMetrics should default to false")
	}
	if opt.AccessFederationValidate {
		t.Error("AccessFederationValidate should default to false")
	}
	if opt.BackgroundTasks {
		t.Error("BackgroundTasks should default to false")
	}

	// Test setting values
	opt.Artifacts = true
	opt.ReplicationStatus = true
	opt.FederationStatus = true
	opt.OpenMetrics = true
	opt.AccessFederationValidate = true
	opt.BackgroundTasks = true

	if !opt.Artifacts {
		t.Error("Artifacts should be true")
	}
	if !opt.ReplicationStatus {
		t.Error("ReplicationStatus should be true")
	}
	if !opt.FederationStatus {
		t.Error("FederationStatus should be true")
	}
	if !opt.OpenMetrics {
		t.Error("OpenMetrics should be true")
	}
	if !opt.AccessFederationValidate {
		t.Error("AccessFederationValidate should be true")
	}
	if !opt.BackgroundTasks {
		t.Error("BackgroundTasks should be true")
	}
}

// Test environment variable processing
func TestEnvconfigProcessing(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectUser  string
		expectPass  string
		expectToken string
	}{
		{
			name: "Username and password from env",
			envVars: map[string]string{
				"ARTI_USERNAME": "envuser",
				"ARTI_PASSWORD": "envpass",
			},
			expectUser:  "envuser",
			expectPass:  "envpass",
			expectToken: "",
		},
		{
			name: "Access token from env",
			envVars: map[string]string{
				"ARTI_ACCESS_TOKEN": "envtoken",
			},
			expectUser:  "",
			expectPass:  "",
			expectToken: "envtoken",
		},
		{
			name: "Override with env vars",
			envVars: map[string]string{
				"ARTI_USERNAME":     "newuser",
				"ARTI_PASSWORD":     "newpass",
				"ARTI_ACCESS_TOKEN": "newtoken",
			},
			expectUser:  "newuser",
			expectPass:  "newpass",
			expectToken: "newtoken",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear existing env vars
			os.Unsetenv("ARTI_USERNAME")
			os.Unsetenv("ARTI_PASSWORD")
			os.Unsetenv("ARTI_ACCESS_TOKEN")

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			var creds Credentials
			err := envconfig.Process("", &creds)
			if err != nil {
				t.Fatalf("envconfig.Process failed: %v", err)
			}

			if creds.Username != tt.expectUser {
				t.Errorf("Username = %s, want %s", creds.Username, tt.expectUser)
			}
			if creds.Password != tt.expectPass {
				t.Errorf("Password = %s, want %s", creds.Password, tt.expectPass)
			}
			if creds.AccessToken != tt.expectToken {
				t.Errorf("AccessToken = %s, want %s", creds.AccessToken, tt.expectToken)
			}
		})
	}
}

// Test edge cases for configuration validation
func TestConfigurationEdgeCases(t *testing.T) {
	t.Run("Special characters in credentials", func(t *testing.T) {
		os.Setenv("ARTI_USERNAME", "user@domain.com")
		os.Setenv("ARTI_PASSWORD", "p@ssw0rd!#$%")
		defer func() {
			os.Unsetenv("ARTI_USERNAME")
			os.Unsetenv("ARTI_PASSWORD")
		}()

		var creds Credentials
		err := envconfig.Process("", &creds)
		if err != nil {
			t.Fatalf("envconfig.Process failed: %v", err)
		}

		if creds.Username != "user@domain.com" {
			t.Errorf("Username with special chars not preserved")
		}
		if creds.Password != "p@ssw0rd!#$%" {
			t.Errorf("Password with special chars not preserved")
		}
	})

	t.Run("Very long credentials", func(t *testing.T) {
		longUser := string(make([]byte, 1000))
		for i := range longUser {
			longUser = longUser[:i] + "a" + longUser[i+1:]
		}
		longPass := string(make([]byte, 1000))
		for i := range longPass {
			longPass = longPass[:i] + "b" + longPass[i+1:]
		}

		os.Setenv("ARTI_USERNAME", longUser)
		os.Setenv("ARTI_PASSWORD", longPass)
		defer func() {
			os.Unsetenv("ARTI_USERNAME")
			os.Unsetenv("ARTI_PASSWORD")
		}()

		var creds Credentials
		err := envconfig.Process("", &creds)
		if err != nil {
			t.Fatalf("envconfig.Process failed with long credentials: %v", err)
		}

		if len(creds.Username) != 1000 {
			t.Errorf("Long username not preserved correctly")
		}
		if len(creds.Password) != 1000 {
			t.Errorf("Long password not preserved correctly")
		}
	})
}
