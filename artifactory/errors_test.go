package artifactory

import (
	"fmt"
	"testing"
)

func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		endpoint string
		status   int
		expected string
	}{
		{
			name:     "Basic API error",
			message:  "Unauthorized",
			endpoint: "/api/system/ping",
			status:   401,
			expected: "API Error: Unauthorized (endpoint: /api/system/ping, status: 401)",
		},
		{
			name:     "Error with no status",
			message:  "Connection refused",
			endpoint: "/api/system/version",
			status:   0,
			expected: "API Error: Connection refused (endpoint: /api/system/version, status: 0)",
		},
		{
			name:     "Empty message",
			message:  "",
			endpoint: "/api/system/license",
			status:   404,
			expected: "API Error:  (endpoint: /api/system/license, status: 404)",
		},
		{
			name:     "Server error",
			message:  "Internal Server Error",
			endpoint: "/api/repositories",
			status:   500,
			expected: "API Error: Internal Server Error (endpoint: /api/repositories, status: 500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &APIError{
				message:  tt.message,
				endpoint: tt.endpoint,
				status:   tt.status,
			}

			if err.Error() != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", err.Error(), tt.expected)
			}

			// Test the apiEndpoint method
			if err.apiEndpoint() != tt.endpoint {
				t.Errorf("APIError.apiEndpoint() = %q, want %q", err.apiEndpoint(), tt.endpoint)
			}

			// Test the apiStatus method
			if err.apiStatus() != tt.status {
				t.Errorf("APIError.apiStatus() = %d, want %d", err.apiStatus(), tt.status)
			}
		})
	}
}

func TestUnmarshalError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		endpoint string
		expected string
	}{
		{
			name:     "Basic unmarshal error",
			message:  "invalid character 'x' looking for beginning of value",
			endpoint: "/api/system/license",
			expected: "Unmarshal Error: invalid character 'x' looking for beginning of value (endpoint: /api/system/license)",
		},
		{
			name:     "Empty message",
			message:  "",
			endpoint: "/api/system/version",
			expected: "Unmarshal Error:  (endpoint: /api/system/version)",
		},
		{
			name:     "Empty endpoint",
			message:  "JSON syntax error",
			endpoint: "",
			expected: "Unmarshal Error: JSON syntax error (endpoint: )",
		},
		{
			name:     "Complex JSON error",
			message:  "cannot unmarshal string into Go struct field BuildInfo.version of type int",
			endpoint: "/api/system/version",
			expected: "Unmarshal Error: cannot unmarshal string into Go struct field BuildInfo.version of type int (endpoint: /api/system/version)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &UnmarshalError{
				message:  tt.message,
				endpoint: tt.endpoint,
			}

			if err.Error() != tt.expected {
				t.Errorf("UnmarshalError.Error() = %q, want %q", err.Error(), tt.expected)
			}

			// Test the apiEndpoint method
			if err.apiEndpoint() != tt.endpoint {
				t.Errorf("UnmarshalError.apiEndpoint() = %q, want %q", err.apiEndpoint(), tt.endpoint)
			}
		})
	}
}

func TestHTTPSuccessCodes(t *testing.T) {
	// Test that our success codes list contains the expected values
	expectedCodes := []int{200, 201, 202, 204}

	for _, code := range expectedCodes {
		t.Run(fmt.Sprintf("Status_%d_should_be_success", code), func(t *testing.T) {
			found := false
			for _, successCode := range httpSuccCodes {
				if successCode == code {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected HTTP status code %d to be in httpSuccCodes", code)
			}
		})
	}

	// Test that error codes are NOT in the success list
	errorCodes := []int{400, 401, 403, 404, 422, 429, 500, 502, 503, 504}

	for _, code := range errorCodes {
		t.Run(fmt.Sprintf("Status_%d_should_not_be_success", code), func(t *testing.T) {
			found := false
			for _, successCode := range httpSuccCodes {
				if successCode == code {
					found = true
					break
				}
			}
			if found {
				t.Errorf("HTTP error status code %d should NOT be in httpSuccCodes", code)
			}
		})
	}

	// Test that we have a reasonable number of success codes
	t.Run("Success_codes_count", func(t *testing.T) {
		if len(httpSuccCodes) == 0 {
			t.Error("httpSuccCodes should not be empty")
		}
		if len(httpSuccCodes) > 10 {
			t.Errorf("httpSuccCodes has %d codes, seems too many for success codes", len(httpSuccCodes))
		}
	})
}

// Test creating errors with various parameters
func TestErrorCreation(t *testing.T) {
	t.Run("APIError_with_all_fields", func(t *testing.T) {
		err := &APIError{
			message:  "Rate limit exceeded",
			endpoint: "/api/storage/repositories",
			status:   429,
		}

		if err.Error() == "" {
			t.Error("APIError.Error() should not return empty string")
		}
		if err.apiEndpoint() != "/api/storage/repositories" {
			t.Error("APIError endpoint not preserved correctly")
		}
		if err.apiStatus() != 429 {
			t.Error("APIError status not preserved correctly")
		}
	})

	t.Run("UnmarshalError_with_all_fields", func(t *testing.T) {
		err := &UnmarshalError{
			message:  "unexpected end of JSON input",
			endpoint: "/api/system/ping",
		}

		if err.Error() == "" {
			t.Error("UnmarshalError.Error() should not return empty string")
		}
		if err.apiEndpoint() != "/api/system/ping" {
			t.Error("UnmarshalError endpoint not preserved correctly")
		}
	})
}

// Test error interface compliance
func TestErrorInterfaceCompliance(t *testing.T) {
	var err error

	t.Run("APIError_implements_error", func(t *testing.T) {
		apiErr := &APIError{message: "test", endpoint: "/test", status: 500}
		err = apiErr
		if err.Error() == "" {
			t.Error("APIError should implement error interface")
		}
	})

	t.Run("UnmarshalError_implements_error", func(t *testing.T) {
		unmarshalErr := &UnmarshalError{message: "test", endpoint: "/test"}
		err = unmarshalErr
		if err.Error() == "" {
			t.Error("UnmarshalError should implement error interface")
		}
	})
}
