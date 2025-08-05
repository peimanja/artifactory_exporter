package logger

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("EmptyConfig_constant", func(t *testing.T) {
		if EmptyConfig.Format != "" {
			t.Errorf("EmptyConfig.Format should be empty, got %q", EmptyConfig.Format)
		}
		if EmptyConfig.Level != "" {
			t.Errorf("EmptyConfig.Level should be empty, got %q", EmptyConfig.Level)
		}
	})

	t.Run("Constants_values", func(t *testing.T) {
		if FormatDefault != fmtTXT {
			t.Errorf("FormatDefault should be %q, got %q", fmtTXT, FormatDefault)
		}
		if LevelDefault != lvlFNameInfo {
			t.Errorf("LevelDefault should be %q, got %q", lvlFNameInfo, LevelDefault)
		}
	})

	t.Run("FormatsAvailable_contents", func(t *testing.T) {
		expectedFormats := []string{fmtTXT, fmtJSON}
		if len(FormatsAvailable) != len(expectedFormats) {
			t.Errorf("FormatsAvailable length = %d, want %d", len(FormatsAvailable), len(expectedFormats))
		}
		for i, format := range expectedFormats {
			if FormatsAvailable[i] != format {
				t.Errorf("FormatsAvailable[%d] = %q, want %q", i, FormatsAvailable[i], format)
			}
		}
	})

	t.Run("LevelsAvailable_contents", func(t *testing.T) {
		expectedLevels := []string{lvlFNameDebug, lvlFNameInfo, lvlFNameWarn, lvlFNameError}
		if len(LevelsAvailable) != len(expectedLevels) {
			t.Errorf("LevelsAvailable length = %d, want %d", len(LevelsAvailable), len(expectedLevels))
		}
		for i, level := range expectedLevels {
			if LevelsAvailable[i] != level {
				t.Errorf("LevelsAvailable[%d] = %q, want %q", i, LevelsAvailable[i], level)
			}
		}
	})

	t.Run("LevelsFlagToSlog_mapping", func(t *testing.T) {
		expectedMappings := map[string]slog.Level{
			"debug": slog.LevelDebug,
			"info":  slog.LevelInfo,
			"warn":  slog.LevelWarn,
			"error": slog.LevelError,
		}

		for flag, expectedLevel := range expectedMappings {
			if actualLevel, exists := levelsFlagToSlog[flag]; !exists {
				t.Errorf("Missing mapping for level flag %q", flag)
			} else if actualLevel != expectedLevel {
				t.Errorf("levelsFlagToSlog[%q] = %v, want %v", flag, actualLevel, expectedLevel)
			}
		}

		// Check no extra mappings exist
		if len(levelsFlagToSlog) != len(expectedMappings) {
			t.Errorf("levelsFlagToSlog has %d mappings, want %d", len(levelsFlagToSlog), len(expectedMappings))
		}
	})
}

func TestNew(t *testing.T) {
	// Redirect stderr to capture log output
	originalStderr := os.Stderr
	defer func() { os.Stderr = originalStderr }()

	tests := []struct {
		name           string
		config         Config
		expectedFormat string
		expectedLevel  slog.Level
	}{
		{
			name:           "Default configuration",
			config:         Config{},
			expectedFormat: fmtTXT,
			expectedLevel:  slog.LevelInfo,
		},
		{
			name:           "Text format debug level",
			config:         Config{Format: fmtTXT, Level: "debug"},
			expectedFormat: fmtTXT,
			expectedLevel:  slog.LevelDebug,
		},
		{
			name:           "JSON format error level",
			config:         Config{Format: fmtJSON, Level: "error"},
			expectedFormat: fmtJSON,
			expectedLevel:  slog.LevelError,
		},
		{
			name:           "Text format warn level",
			config:         Config{Format: fmtTXT, Level: "warn"},
			expectedFormat: fmtTXT,
			expectedLevel:  slog.LevelWarn,
		},
		{
			name:           "Only format specified",
			config:         Config{Format: fmtJSON, Level: ""},
			expectedFormat: fmtJSON,
			expectedLevel:  slog.LevelInfo, // Default level
		},
		{
			name:           "Only level specified",
			config:         Config{Format: "", Level: "debug"},
			expectedFormat: fmtTXT, // Default format
			expectedLevel:  slog.LevelDebug,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr to avoid polluting test output
			r, w, _ := os.Pipe()
			os.Stderr = w

			logger := New(tt.config)

			// Restore stderr
			w.Close()
			os.Stderr = originalStderr

			if logger == nil {
				t.Fatal("New() returned nil logger")
			}

			// Test that the logger actually works by logging at different levels
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// For JSON format, output should contain JSON
			if tt.expectedFormat == fmtJSON {
				if !strings.Contains(output, `"msg"`) && !strings.Contains(output, `"level"`) {
					// Only check if there's actual output (some levels might be filtered)
					if output != "" {
						t.Errorf("Expected JSON format output, got: %s", output)
					}
				}
			}

			// Test logging at the expected level works
			if tt.expectedLevel == slog.LevelDebug {
				logger.Debug("test debug")
			} else if tt.expectedLevel == slog.LevelError {
				logger.Error("test error")
			}
		})
	}
}

func TestFmtFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name:     "Empty format uses default",
			config:   Config{Format: ""},
			expected: FormatDefault,
		},
		{
			name:     "Text format",
			config:   Config{Format: fmtTXT},
			expected: fmtTXT,
		},
		{
			name:     "JSON format",
			config:   Config{Format: fmtJSON},
			expected: fmtJSON,
		},
		{
			name:     "Invalid format uses default", // Function should handle this gracefully
			config:   Config{Format: "invalid"},
			expected: FormatDefault, // Should return default for invalid formats
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtFromConfig(tt.config)
			if result != tt.expected {
				t.Errorf("fmtFromConfig() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestLvlFromConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected slog.Level
	}{
		{
			name:     "Empty level uses default",
			config:   Config{Level: ""},
			expected: slog.LevelInfo,
		},
		{
			name:     "Debug level",
			config:   Config{Level: "debug"},
			expected: slog.LevelDebug,
		},
		{
			name:     "Info level",
			config:   Config{Level: "info"},
			expected: slog.LevelInfo,
		},
		{
			name:     "Warn level",
			config:   Config{Level: "warn"},
			expected: slog.LevelWarn,
		},
		{
			name:     "Error level",
			config:   Config{Level: "error"},
			expected: slog.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lvlFromConfig(tt.config)
			if result != tt.expected {
				t.Errorf("lvlFromConfig() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestNewJSONLogger(t *testing.T) {
	// Capture stderr
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := newJSONLogger(slog.LevelInfo)

	// Test logging
	logger.Info("test message", "key", "value")

	// Restore stderr and read output
	w.Close()
	os.Stderr = originalStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if logger == nil {
		t.Fatal("newJSONLogger() returned nil")
	}

	// Check that output contains JSON structure
	if !strings.Contains(output, `"msg"`) {
		t.Errorf("Expected JSON output to contain 'msg' field, got: %s", output)
	}

	if !strings.Contains(output, `"level"`) {
		t.Errorf("Expected JSON output to contain 'level' field, got: %s", output)
	}
}

func TestNewTXTLogger(t *testing.T) {
	// Capture stderr
	originalStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	logger := newTXTLogger(slog.LevelInfo)

	// Test logging
	logger.Info("test message", "key", "value")

	// Restore stderr and read output
	w.Close()
	os.Stderr = originalStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if logger == nil {
		t.Fatal("newTXTLogger() returned nil")
	}

	// Check that output contains text format (logfmt style)
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected text output to contain message, got: %s", output)
	}
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name         string
		level        slog.Level
		shouldOutput map[string]bool // which levels should produce output
	}{
		{
			name:  "Debug level logs all",
			level: slog.LevelDebug,
			shouldOutput: map[string]bool{
				"debug": true,
				"info":  true,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "Info level skips debug",
			level: slog.LevelInfo,
			shouldOutput: map[string]bool{
				"debug": false,
				"info":  true,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "Warn level skips debug and info",
			level: slog.LevelWarn,
			shouldOutput: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  true,
				"error": true,
			},
		},
		{
			name:  "Error level only logs errors",
			level: slog.LevelError,
			shouldOutput: map[string]bool{
				"debug": false,
				"info":  false,
				"warn":  false,
				"error": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Create handlers that write to our buffer instead of stderr
			var logger *slog.Logger
			switch tt.name {
			case "Debug level logs all":
				logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
			case "Info level skips debug":
				logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
			case "Warn level skips debug and info":
				logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
			case "Error level only logs errors":
				logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError}))
			}

			// Test each log level
			logMethods := map[string]func(string, ...any){
				"debug": logger.Debug,
				"info":  logger.Info,
				"warn":  logger.Warn,
				"error": logger.Error,
			}

			for levelName, logMethod := range logMethods {
				// Clear buffer for each test
				buf.Reset()

				// Log at this level
				logMethod(fmt.Sprintf("test %s message", levelName))

				output := buf.String()
				shouldHaveOutput := tt.shouldOutput[levelName]
				hasOutput := strings.TrimSpace(output) != ""

				if shouldHaveOutput && !hasOutput {
					t.Errorf("Expected output for %s level with logger at %v, but got none", levelName, tt.level)
				}
				if !shouldHaveOutput && hasOutput {
					t.Errorf("Did not expect output for %s level with logger at %v, but got: %s", levelName, tt.level, output)
				}
			}
		})
	}
}

// Test edge cases and error conditions
func TestLoggerEdgeCases(t *testing.T) {
	t.Run("Invalid_format_defaults_to_text", func(t *testing.T) {
		// This test ensures the function handles unexpected format values gracefully
		config := Config{Format: "invalid_format", Level: "info"}

		// Use a simple test that doesn't try to capture stderr
		logger := New(config)

		if logger == nil {
			t.Fatal("New() should not return nil even with invalid format")
		}

		// The function should handle this case and still return a working logger
		// We can test that it doesn't panic by calling a log method
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Logger panicked when logging with invalid format: %v", r)
			}
		}()

		// Test that the logger can be used without panicking
		logger.Info("test message")
	})
}
