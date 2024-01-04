package artifactory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	msgErrAPICall = "There was an error making API call" // https://github.com/peimanja/artifactory_exporter/pull/121/checks?check_run_id=20136336585
)

// APIErrors represents Artifactory API Error response
type APIErrors struct {
	Errors interface{}
}

type ApiResponse struct {
	Body   []byte
	NodeId string
}

func (c *Client) makeRequest(method string, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, path, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error(
			"There was an error creating request",
			"err", err.Error(),
		)
		return nil, err
	}
	switch c.authMethod {
	case "userPass":
		req.SetBasicAuth(c.cred.Username, c.cred.Password)
	case "accessToken":
		req.Header.Add("Authorization", "Bearer "+c.cred.AccessToken)
	default:
		return nil, fmt.Errorf("Artifactory Auth (%s) method is not supported", c.authMethod)
	}
	return c.client.Do(req)
}

// FetchHTTP is a wrapper function for making all Get API calls
func (c *Client) FetchHTTP(path string) (*ApiResponse, error) {
	var response ApiResponse
	fullPath := fmt.Sprintf("%s/api/%s", c.URI, path)
	c.logger.Debug(
		"Fetching http",
		"path", fullPath,
	)
	resp, err := c.makeRequest("GET", fullPath, nil)
	if err != nil {
		c.logger.Error(
			msgErrAPICall,
			"endpoint", fullPath,
			"err", err.Error(),
		)
		return nil, err
	}
	response.NodeId = resp.Header.Get("x-artifactory-node-id")
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			c.logger.Error(
				"There was an error when trying to unmarshal the API Error",
				"err", err,
			)
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		c.logger.Warn(
			"The endpoint does not exist",
			"endpoint", fullPath,
			"err", fmt.Sprintf("%v", apiErrors.Errors),
			"status", 404,
		)
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
			status:   404,
		}
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			c.logger.Error(
				"There was an error when trying to unmarshal the API Error",
				"err", err.Error(),
			)
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		c.logger.Error(
			msgErrAPICall,
			"endpoint", fullPath,
			"err", fmt.Sprintf("%v", apiErrors.Errors),
			"status", "is missing and should be provided", // Value from Kacper Perschke and should be changed to significant!
		)
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(
			"There was an error reading response body",
			"err", err.Error(),
		)
		return nil, err
	}
	response.Body = bodyBytes

	return &response, nil
}

// QueryAQL is a wrapper function for making an query to AQL endpoint
func (c *Client) QueryAQL(query []byte) (*ApiResponse, error) {
	var response ApiResponse
	fullPath := fmt.Sprintf("%s/api/search/aql", c.URI)
	c.logger.Debug(
		"Running AQL query",
		"path", fullPath,
	)
	resp, err := c.makeRequest("POST", fullPath, query)
	if err != nil {
		c.logger.Error(
			msgErrAPICall,
			"endpoint", fullPath,
			"err", err.Error(),
		)
		return nil, err
	}
	response.NodeId = resp.Header.Get("x-artifactory-node-id")
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		c.logger.Error(
			msgErrAPICall,
			"endpoint", fullPath,
			"err", fmt.Sprintf("%v", apiErrors.Errors),
			"status", "evaporated?", // Value from Kacper Perschke and should be changed to significant!
		)
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(
			"There was an error reading response body",
			"err", err.Error(),
		)
		return nil, err
	}
	response.Body = bodyBytes
	return &response, nil
}
