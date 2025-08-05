package artifactory

import (
	"encoding/json"
	"testing"
)

func TestLicenseInfo_JSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expected    LicenseInfo
		expectError bool
	}{
		{
			name:      "Standard Enterprise license",
			jsonInput: `{"type":"Enterprise","validThrough":"Dec 31, 2025","licensedTo":"Test Company"}`,
			expected: LicenseInfo{
				Type:         "Enterprise",
				ValidThrough: "Dec 31, 2025",
				LicensedTo:   "Test Company",
			},
			expectError: false,
		},
		{
			name:      "OSS license",
			jsonInput: `{"type":"OSS","validThrough":"","licensedTo":""}`,
			expected: LicenseInfo{
				Type:         "OSS",
				ValidThrough: "",
				LicensedTo:   "",
			},
			expectError: false,
		},
		{
			name:      "N/R validThrough",
			jsonInput: `{"type":"Pro","validThrough":"N/R","licensedTo":"Test Company"}`,
			expected: LicenseInfo{
				Type:         "Pro",
				ValidThrough: "N/R",
				LicensedTo:   "Test Company",
			},
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			jsonInput:   `{"type":"Enterprise","validThrough":}`,
			expected:    LicenseInfo{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var license LicenseInfo
			err := json.Unmarshal([]byte(tt.jsonInput), &license)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error unmarshaling JSON, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error unmarshaling JSON: %v", err)
				return
			}

			if license.Type != tt.expected.Type {
				t.Errorf("Type = %v, want %v", license.Type, tt.expected.Type)
			}
			if license.ValidThrough != tt.expected.ValidThrough {
				t.Errorf("ValidThrough = %v, want %v", license.ValidThrough, tt.expected.ValidThrough)
			}
			if license.LicensedTo != tt.expected.LicensedTo {
				t.Errorf("LicensedTo = %v, want %v", license.LicensedTo, tt.expected.LicensedTo)
			}
		})
	}
}

func TestBuildInfo_JSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name        string
		jsonInput   string
		expected    BuildInfo
		expectError bool
	}{
		{
			name: "Complete build info",
			jsonInput: `{
				"version": "7.41.7",
				"revision": "74107900",
				"addons": ["ha", "enterprise"],
				"license": "Enterprise"
			}`,
			expected: BuildInfo{
				Version:  "7.41.7",
				Revision: "74107900",
				Addons:   []string{"ha", "enterprise"},
				License:  "Enterprise",
			},
			expectError: false,
		},
		{
			name: "Minimal build info",
			jsonInput: `{
				"version": "7.41.7",
				"revision": "74107900"
			}`,
			expected: BuildInfo{
				Version:  "7.41.7",
				Revision: "74107900",
				Addons:   nil,
				License:  "",
			},
			expectError: false,
		},
		{
			name:        "Invalid JSON",
			jsonInput:   `{"version":}`,
			expected:    BuildInfo{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buildInfo BuildInfo
			err := json.Unmarshal([]byte(tt.jsonInput), &buildInfo)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error unmarshaling JSON, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error unmarshaling JSON: %v", err)
				return
			}

			if buildInfo.Version != tt.expected.Version {
				t.Errorf("Version = %v, want %v", buildInfo.Version, tt.expected.Version)
			}
			if buildInfo.Revision != tt.expected.Revision {
				t.Errorf("Revision = %v, want %v", buildInfo.Revision, tt.expected.Revision)
			}
			if buildInfo.License != tt.expected.License {
				t.Errorf("License = %v, want %v", buildInfo.License, tt.expected.License)
			}

			// Check addons slice
			if len(buildInfo.Addons) != len(tt.expected.Addons) {
				t.Errorf("Addons length = %v, want %v", len(buildInfo.Addons), len(tt.expected.Addons))
			} else {
				for i, addon := range buildInfo.Addons {
					if addon != tt.expected.Addons[i] {
						t.Errorf("Addons[%d] = %v, want %v", i, addon, tt.expected.Addons[i])
					}
				}
			}
		})
	}
}

// Test the date normalization logic specifically
func TestValidThroughDateNormalization(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
		desc      string
	}{
		{
			name:      "XX day gets normalized",
			input:     "Jul XX, 2025",
			expectErr: false,
			desc:      "Should normalize XX to 1",
		},
		{
			name:      "Regular day unchanged",
			input:     "Jul 15, 2025",
			expectErr: false,
			desc:      "Should keep numeric day as-is",
		},
		{
			name:      "Single digit day",
			input:     "Jan 5, 2026",
			expectErr: false,
			desc:      "Should handle single digit days",
		},
		{
			name:      "Two digit day",
			input:     "Dec 25, 2025",
			expectErr: false,
			desc:      "Should handle two digit days",
		},
		{
			name:      "Invalid format",
			input:     "Not a date",
			expectErr: true,
			desc:      "Should fail on invalid format",
		},
		{
			name:      "N/R format",
			input:     "N/R",
			expectErr: true,
			desc:      "Should fail on N/R format",
		},
		{
			name:      "Empty string",
			input:     "",
			expectErr: true,
			desc:      "Should fail on empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := LicenseInfo{
				Type:         "Enterprise", // Non-OSS to trigger parsing
				ValidThrough: tt.input,
			}

			_, err := license.ValidSeconds()

			if tt.expectErr && err == nil {
				t.Errorf("Expected error for input '%s', but got none", tt.input)
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error for input '%s': %v", tt.input, err)
			}

			t.Logf("%s: %s", tt.desc, tt.input)
		})
	}
}
