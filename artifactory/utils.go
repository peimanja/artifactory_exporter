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
	Errors interface{}
}

// FetchHTTP is a wrapper function for making all Get API calls
func (c *Client) FetchHTTP(path string) ([]byte, error) {
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

	if resp.StatusCode == 404 {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			level.Error(c.logger).Log("msg", "There was an error when trying to unmarshal the API Error", "err", err)
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		level.Warn(c.logger).Log("msg", "The endpoint does not exist", "endpoint", fullPath, "err", fmt.Sprintf("%v", apiErrors.Errors), "status", 404)
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
			level.Error(c.logger).Log("msg", "There was an error when trying to unmarshal the API Error", "err", err)
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		level.Error(c.logger).Log("msg", "There was an error making API call", "endpoint", fullPath, "err", fmt.Sprintf("%v", apiErrors.Errors), "status")
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

// QueryAQL is a wrapper function for making an query to AQL endpoint
func (c *Client) QueryAQL(query []byte) ([]byte, error) {
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
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		level.Error(c.logger).Log("msg", "There was an error making API call", "endpoint", fullPath, "err", fmt.Sprintf("%v", apiErrors.Errors), "status")
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
		}
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}
