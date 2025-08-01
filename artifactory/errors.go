package artifactory

//UnmarshalError is a custom Error type for unmarshal API respond body error
type UnmarshalError struct {
	message  string
	endpoint string
}

func (e *UnmarshalError) Error() string {
	return e.message
}

func (e *UnmarshalError) apiEndpoint() string {
	return e.endpoint
}

//APIError is a custom Error type for API error
type APIError struct {
	message  string
	endpoint string
	status   int
}

func (e *APIError) Error() string {
	return e.message
}

func (e *APIError) apiEndpoint() string {
	return e.endpoint
}

func (e *APIError) apiStatus() int {
	return e.status
}

// Status returns the HTTP status code of the API error
func (e *APIError) Status() int {
	return e.status
}
