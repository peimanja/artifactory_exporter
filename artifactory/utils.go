package artifactory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log/level"
)

// APIErrors represents Artifactory API Error response
type APIErrors struct {
	Errors []struct {
		Status  int    `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
}

func (c *Client) fetchHTTP(path string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/api/%s", c.URI, path)
	level.Debug(c.logger).Log("msg", "Fetching http", "path", fullPath)
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	switch c.authMethod {
	case "userPass":
		req.SetBasicAuth(c.cred.Username, c.cred.Password)
	case "accessToken":
		req.Header.Add("Authorization", "Bearer "+c.cred.AccessToken)
	default:
		return nil, fmt.Errorf("Artifactory Auth method is not supported")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &apiErrors)
		level.Error(c.logger).Log("msg", "There was an error making API call", "endpoint", fullPath, "err", apiErrors.Errors[0].Message, "status", apiErrors.Errors[0].Status)
		return nil, &APIError{
			message:  apiErrors.Errors[0].Message,
			endpoint: fullPath,
			status:   apiErrors.Errors[0].Status,
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func (c *Client) queryAQL(query []byte) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/api/search/aql", c.URI)
	level.Debug(c.logger).Log("msg", "Running AQL query", "path", fullPath)
	req, err := http.NewRequest("POST", fullPath, bytes.NewBuffer(query))
	req.Header = http.Header{"Content-Type": {"text/plain"}}
	if err != nil {
		return nil, err
	}
	switch c.authMethod {
	case "userPass":
		req.SetBasicAuth(c.cred.Username, c.cred.Password)
	case "accessToken":
		req.Header.Add("Authorization", "Bearer "+c.cred.AccessToken)
	default:
		return nil, fmt.Errorf("Artifactory Auth method is not supported")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &apiErrors)
		level.Error(c.logger).Log("msg", "There was an error making API call", "endpoint", fullPath, "err", apiErrors.Errors[0].Message, "status", apiErrors.Errors[0].Status)
		return nil, &APIError{
			message:  apiErrors.Errors[0].Message,
			endpoint: fullPath,
			status:   apiErrors.Errors[0].Status,
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
