package collector

import (
	"log/slog"
	"math"
	"regexp"
	"testing"

	l "github.com/peimanja/artifactory_exporter/logger"
)

const float64EqualityThreshold = 1e-6

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

// newTestLogger creates a logger instance for testing
func newTestLogger() *slog.Logger {
	return l.New(l.Config{
		Format: l.FormatDefault,
		Level:  "debug",
	})
}

var testExporter = &Exporter{
	logger: newTestLogger(),
}

func TestConvMultiplier(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{"Percentage", "%", 0.01, false},
		{"Bytes", "bytes", 1, false},
		{"Kilobytes", "KB", 1024, false},
		{"Megabytes", "MB", 1048576, false},
		{"Gigabytes", "GB", 1073741824, false},
		{"Terabytes", "TB", 1099511627776, false},
		{"Invalid unit", "XB", 0, true},
		{"Empty string", "", 0, true},
		{"Lowercase kb", "kb", 0, true}, // Should fail - case sensitive
		{"Case sensitive MB", "mb", 0, true},
		{"Case sensitive TB", "tb", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testExporter.convMultiplier(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("convMultiplier('%s') = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvNumber(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{"Integer", "123", 123.0, false},
		{"Float", "123.45", 123.45, false},
		{"Zero", "0", 0.0, false},
		{"Negative", "-123.45", -123.45, false},
		{"Large number", "1234567890.123", 1234567890.123, false},
		{"Scientific notation", "1.23e10", 1.23e10, false},
		{"Very small number", "0.0001", 0.0001, false},
		{"Number with comma", "1,234", 0, true}, // Should fail - not valid float
		{"Invalid number", "abc", 0, true},
		{"Empty string", "", 0, true},
		{"Mixed alphanumeric", "123abc", 0, true},
		{"Special characters", "123.45.67", 0, true},
		{"Infinity", "+Inf", math.Inf(1), false},
		{"Negative infinity", "-Inf", math.Inf(-1), false},
		{"NaN", "NaN", math.NaN(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testExporter.convNumber(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			// Special handling for NaN
			if math.IsNaN(tt.expected) {
				if !math.IsNaN(result) {
					t.Errorf("convNumber('%s') = %f, want NaN", tt.input, result)
				}
				return
			}

			if result != tt.expected {
				t.Errorf("convNumber('%s') = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractNamedGroups(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		input    string
		expected map[string]string
	}{
		{
			name:    "Simple named groups",
			pattern: `^(?P<number>\d+) (?P<unit>[A-Z]+)$`,
			input:   "123 KB",
			expected: map[string]string{
				"number": "123",
				"unit":   "KB",
			},
		},
		{
			name:    "File store pattern",
			pattern: `^(?P<size>[[:digit:]]{1,3}(?:[[:digit:]]|(?:,[[:digit:]]{3})*(?:\.[[:digit:]]{1,2})?)?) [KMGT]B \((?P<usage>[[:digit:]]{1,2}(?:\.[[:digit:]]{1,2})?%)\)$`,
			input:   "3.33 TB (3.3%)",
			expected: map[string]string{
				"size":  "3.33",
				"usage": "3.3%",
			},
		},
		{
			name:     "No match",
			pattern:  `^(?P<number>\d+)$`,
			input:    "abc",
			expected: map[string]string{},
		},
		{
			name:     "Empty groups",
			pattern:  `^(?P<number>\d*)(?P<unit>[A-Z]*)$`,
			input:    "",
			expected: map[string]string{},
		},
		{
			name:    "Multiple groups with some empty",
			pattern: `^(?P<first>\w*) (?P<second>\w*) (?P<third>\w*)$`,
			input:   "hello  world", // Double space means second group will be empty
			expected: map[string]string{
				"first": "hello",
				"third": "world",
			},
		},
		{
			name:    "Groups with special characters",
			pattern: `^(?P<amount>[\d,]+\.?\d*) (?P<unit>[A-Z]+) \((?P<percent>[\d.]+%)\)$`,
			input:   "1,234.56 GB (78.9%)",
			expected: map[string]string{
				"amount":  "1,234.56",
				"unit":    "GB",
				"percent": "78.9%",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := regexp.MustCompile(tt.pattern)
			result := extractNamedGroups(tt.input, re)

			if len(result) != len(tt.expected) {
				t.Errorf("extractNamedGroups() returned %d groups, expected %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected group '%s' not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Group '%s' = '%s', want '%s'", key, actualValue, expectedValue)
				}
			}

			// Check that no unexpected groups exist
			for key := range result {
				if _, expected := tt.expected[key]; !expected {
					t.Errorf("Unexpected group '%s' found with value '%s'", key, result[key])
				}
			}
		})
	}
}

func TestConvArtiToPromBool(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected float64
	}{
		{"True converts to 1", true, 1.0},
		{"False converts to 0", false, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convArtiToPromBool(tt.input)
			if result != tt.expected {
				t.Errorf("convArtiToPromBool(%t) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

// Test the complete conversion pipeline with realistic values
func TestConvArtiToPromNumber(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    float64
		expectError bool
	}{
		{"Plain bytes", "8 bytes", 8.0, false},
		{"Bytes without space", "8bytes", 8.0, false},
		{"Megabytes with comma", "8,888.88 MB", 9320666234.879999, false},
		{"Gigabytes", "88.88 GB", 95434173317.119995, false},
		{"Large gigabytes", "888.88 GB", 954427632517.119995, false},
		{"Plain number", "1", 1.0, false},
		{"Large plain number", "44", 44.0, false},
		{"Percentage", "100 %", 1.0, false},
		{"Gigabytes no space", "1000GB", 1073741824000.0, false},
		{"Kilobytes", "9999 KB", 10238976.0, false},
		{"Number with commas", "32,564,943", 32564943.0, false},
		{"Terabytes", "1.5 TB", 1649267441664.0, false},
		{"Percentage small", "0.5%", 0.005, false},
		{"Zero bytes", "0 bytes", 0.0, false},
		{"Invalid format", "abc GB", 0.0, true},
		{"Unknown unit", "100 PB", 0.0, true},
		{"Empty string", "", 0.0, true},
		{"Just unit", "GB", 0.0, true},
		{"Just number", "123.45", 123.45, false}, // Plain number should work
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := testExporter.convArtiToPromNumber(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if !almostEqual(result, tt.expected) {
				t.Errorf("convArtiToPromNumber('%s') = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvArtiToPromFileStoreData(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedSize  float64
		expectedUsage float64
		expectError   bool
	}{
		{"TB with percentage", "3.33 TB (3.3%)", 3661373720494.080078, 0.033, false},
		{"TB with higher percentage", "6.66 TB (6.66%)", 7322747440988.160156, 0.0666, false},
		{"Large TB", "11.11 TB (11.1%)", 12215574184591.359375, 0.111, false},
		{"Very large TB", "99.99 TB (99.99%)", 109940167661322.234375, 0.9999, false},
		{"GB without percentage", "499.76 GB", 536613213962.23999, 0.0, false},
		{"GB with small percentage", "4.82 GB (0.96%)", 5175435591.68, 0.0096, false},
		{"GB with high percentage", "494.94 GB (99.04%)", 531437778370.559998, 0.9904, false},
		{"Round numbers", "1.0 GB (1.0%)", 1073741824.0, 0.01, false},
		{"GB with comma", "1,427.32 GB (18.2%)", 1532573180231.68, 0.182, false},
		{"MB values", "512.5 MB (50.0%)", 537395200.0, 0.5, false},
		{"KB values", "1024 KB (25.5%)", 1048576.0, 0.255, false},
		{"Zero size", "0 GB", 0.0, 0.0, false},
		{"100 percent", "1 TB (100%)", 1099511627776.0, 1.0, false},
		{"Invalid format", "invalid", 0.0, 0.0, true},
		{"Missing percentage", "invalid GB format", 0.0, 0.0, true},
		{"Invalid unit", "100 PB (50%)", 0.0, 0.0, true},
		{"S3 sharding N/A format", "0 bytes (N/A)", 0.0, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, usage, err := testExporter.convArtiToPromFileStoreData(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
				return
			}

			if !almostEqual(size, tt.expectedSize) {
				t.Errorf("Size: convArtiToPromFileStoreData('%s') = %f, want %f", tt.input, size, tt.expectedSize)
			}

			if !almostEqual(usage, tt.expectedUsage) {
				t.Errorf("Usage: convArtiToPromFileStoreData('%s') = %f, want %f", tt.input, usage, tt.expectedUsage)
			}
		})
	}
}

func TestConvArtiToPromFileStoreData_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Malformed input missing space after size",
			input:       "100GB(50%)", // Missing space before unit
			expectError: true,
			errorMsg:    "does not match",
		},
		{
			name:        "Empty size in matched regex",
			input:       " TB (50%)", // Empty size part
			expectError: true,
			errorMsg:    "does not match",
		},
		{
			name:        "Invalid percentage over 100",
			input:       "100 GB (150%)", // Over 100%
			expectError: true,
			errorMsg:    "does not match",
		},
		{
			name:        "Invalid percentage format",
			input:       "100 GB (abc%)", // Non-numeric percentage
			expectError: true,
			errorMsg:    "does not match",
		},
		{
			name:        "Missing closing parenthesis",
			input:       "100 GB (50%", // Missing closing paren
			expectError: true,
			errorMsg:    "does not match",
		},
		{
			name:        "Valid edge case 100%",
			input:       "100 GB (100%)",
			expectError: false,
		},
		{
			name:        "Valid edge case 0%",
			input:       "100 GB (0%)",
			expectError: false,
		},
		{
			name:        "Valid edge case with decimals",
			input:       "100 GB (99.99%)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := testExporter.convArtiToPromFileStoreData(tt.input)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for input '%s', but got none", tt.input)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for input '%s', but got: %v", tt.input, err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !regexp.MustCompile(tt.errorMsg).MatchString(err.Error()) {
					t.Errorf("Expected error message to contain '%s', but got: %v", tt.errorMsg, err.Error())
				}
			}
		})
	}
}
