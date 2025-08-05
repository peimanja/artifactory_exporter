package artifactory

import "fmt"

// UnmarshalError is a custom Error type for unmarshal API respond body error
type UnmarshalError struct {
	message  string
	endpoint string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("Unmarshal Error: %s (endpoint: %s)", e.message, e.endpoint)
}

func (e *UnmarshalError) apiEndpoint() string {
	return e.endpoint
}

// APIError is a custom Error type for API error
type APIError struct {
	message  string
	endpoint string
	status   int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %s (endpoint: %s, status: %d)", e.message, e.endpoint, e.status)
}

func (e *APIError) apiEndpoint() string {
	return e.endpoint
}

func (e *APIError) apiStatus() int {
	return e.status
}
