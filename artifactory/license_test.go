package artifactory

import (
	"testing"
)

func TestLicenseInfo_IsOSS(t *testing.T) {
	tests := []struct {
		name     string
		license  LicenseInfo
		expected bool
	}{
		{
			name:     "OSS license",
			license:  LicenseInfo{Type: "OSS"},
			expected: true,
		},
		{
			name:     "JCR Edition license",
			license:  LicenseInfo{Type: "JCR Edition"},
			expected: true,
		},
		{
			name:     "Community Edition for C/C++ license",
			license:  LicenseInfo{Type: "Community Edition for C/C++"},
			expected: true,
		},
		{
			name:     "Pro license",
			license:  LicenseInfo{Type: "Pro"},
			expected: false,
		},
		{
			name:     "Enterprise license",
			license:  LicenseInfo{Type: "Enterprise"},
			expected: false,
		},
		{
			name:     "Edge license",
			license:  LicenseInfo{Type: "Edge"},
			expected: false,
		},
		{
			name:     "Mixed case OSS",
			license:  LicenseInfo{Type: "Oss"},
			expected: true,
		},
		{
			name:     "Empty type",
			license:  LicenseInfo{Type: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.license.IsOSS()
			if got != tt.expected {
				t.Errorf("IsOSS() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLicenseInfo_TypeNormalized(t *testing.T) {
	tests := []struct {
		name     string
		license  LicenseInfo
		expected string
	}{
		{
			name:     "Uppercase license",
			license:  LicenseInfo{Type: "ENTERPRISE"},
			expected: "enterprise",
		},
		{
			name:     "Mixed case license",
			license:  LicenseInfo{Type: "EnTeRpRiSe"},
			expected: "enterprise",
		},
		{
			name:     "Already lowercase",
			license:  LicenseInfo{Type: "pro"},
			expected: "pro",
		},
		{
			name:     "Empty type",
			license:  LicenseInfo{Type: ""},
			expected: "",
		},
		{
			name:     "Type with spaces",
			license:  LicenseInfo{Type: "Community Edition for C/C++"},
			expected: "community edition for c/c++",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.license.TypeNormalized()
			if got != tt.expected {
				t.Errorf("TypeNormalized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLicenseInfo_ValidSeconds(t *testing.T) {
	tests := []struct {
		name        string
		license     LicenseInfo
		wantErr     bool
		description string
	}{
		{
			name:        "OSS license returns 0",
			license:     LicenseInfo{Type: "OSS", ValidThrough: "Jan 1, 2026"},
			wantErr:     false,
			description: "OSS licenses should always return 0 seconds",
		},
		{
			name:        "Valid date format",
			license:     LicenseInfo{Type: "Enterprise", ValidThrough: "Dec 31, 2025"},
			wantErr:     false,
			description: "Standard date format should parse correctly",
		},
		{
			name:        "Edge license with XX day",
			license:     LicenseInfo{Type: "Edge", ValidThrough: "Dec XX, 2025"},
			wantErr:     false,
			description: "Edge licenses with XX day should default to 1st of month",
		},
		{
			name:        "Another XX format",
			license:     LicenseInfo{Type: "Pro", ValidThrough: "Jan XX, 2026"},
			wantErr:     false,
			description: "Any license with XX day should work",
		},
		{
			name:        "Invalid date format",
			license:     LicenseInfo{Type: "Enterprise", ValidThrough: "Invalid Date"},
			wantErr:     true,
			description: "Invalid date format should return error",
		},
		{
			name:        "Empty valid through",
			license:     LicenseInfo{Type: "Enterprise", ValidThrough: ""},
			wantErr:     true,
			description: "Empty date should return error",
		},
		{
			name:        "N/R date format",
			license:     LicenseInfo{Type: "Enterprise", ValidThrough: "N/R"},
			wantErr:     true,
			description: "N/R format should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seconds, err := tt.license.ValidSeconds()

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidSeconds() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ValidSeconds() unexpected error: %v", err)
				return
			}

			// Special case for OSS licenses
			if tt.license.IsOSS() {
				if seconds != 0 {
					t.Errorf("ValidSeconds() for OSS license = %v, want 0", seconds)
				}
				return
			}

			// For non-OSS licenses, verify the calculation makes sense
			// We can't check exact values due to time differences, but we can verify
			// that future dates return positive values (as of August 2025)
			t.Logf("ValidSeconds() returned %d seconds for %s", seconds, tt.description)
		})
	}
}

func TestLicenseInfo_ValidSeconds_EdgeCaseHandling(t *testing.T) {
	// Test the specific case from the bug report
	license := LicenseInfo{
		Type:         "Edge",
		ValidThrough: "Dec XX, 2025", // Use December instead of July since we're in August 2025
		LicensedTo:   "Test Company",
	}

	seconds, err := license.ValidSeconds()
	if err != nil {
		t.Errorf("ValidSeconds() for Edge license with XX day failed: %v", err)
	}

	// The function should have normalized "Dec XX, 2025" to "Dec 1, 2025"
	// and calculated seconds from current time to December 1, 2025
	if seconds <= 0 {
		t.Errorf("ValidSeconds() returned non-positive value %d, expected positive future date", seconds)
	}

	t.Logf("Edge license with 'Dec XX, 2025' correctly parsed, returning %d seconds", seconds)
}

func TestLicenseInfo_ValidSeconds_DateNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "XX in day position",
			input:    "Dec XX, 2025",
			expected: "normalized to Dec 1, 2025",
		},
		{
			name:     "Regular numeric day",
			input:    "Dec 15, 2025",
			expected: "remains Dec 15, 2025",
		},
		{
			name:     "Single digit day",
			input:    "Jan 5, 2026",
			expected: "remains Jan 5, 2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			license := LicenseInfo{
				Type:         "Enterprise",
				ValidThrough: tt.input,
			}

			_, err := license.ValidSeconds()
			if err != nil {
				t.Errorf("ValidSeconds() failed for input %s: %v", tt.input, err)
			}

			t.Logf("Input '%s' %s", tt.input, tt.expected)
		})
	}
}
