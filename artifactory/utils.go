package artifactory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"slices"
	"strings"
)

const (
	logMsgErrAPICall    = "There was an error making API call"
	logMsgErrUnmarshall = "There was an error when trying to unmarshal the API Error"
)

// APIErrors represents Artifactory API Error response
type APIErrors struct {
	Errors interface{}
}

type ApiResponse struct {
	Body   []byte
	NodeId string
}

var (
	httpSuccCodes = []int{ // https://go.dev/src/net/http/status.go
		http.StatusOK,                   // 200
		http.StatusCreated,              // 201
		http.StatusAccepted,             // 202
		http.StatusNonAuthoritativeInfo, // 203
		http.StatusNoContent,            // 204
		http.StatusResetContent,         // 205
		http.StatusPartialContent,       // 206
		http.StatusMultiStatus,          // 207
		http.StatusAlreadyReported,      // 208
		http.StatusIMUsed,               // 226
	}
)

func (c *Client) makeRequest(method string, path string, body []byte, headers **map[string]string) (*http.Response, error) {
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
	if headers != nil {
		for key, value := range **headers {
			req.Header.Set(key, value)
		}
	}
	return c.client.Do(req)
}

func (c *Client) procRespErr(resp *http.Response, fPath string) (*ApiResponse, error) {
	var apiErrors APIErrors
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
		c.logger.Error(
			logMsgErrUnmarshall,
			"err", err.Error(),
		)
		return nil, &UnmarshalError{
			message:  err.Error(),
			endpoint: fPath,
		}
	}
	c.logger.Error(
		logMsgErrAPICall,
		"endpoint", fPath,
		"err", fmt.Sprintf("%v", apiErrors.Errors),
		"status", resp.StatusCode,
	)
	return nil, &APIError{
		message:  fmt.Sprintf("%v", apiErrors.Errors),
		endpoint: fPath,
		// status:   resp.StatusCode, // Maybe it would be worth returning it too? As with http.StatusNotFound.
	}
}

// FetchHTTP is a wrapper function for making all Get API calls
func (c *Client) FetchHTTP(path string) (*ApiResponse, error) {
	var response ApiResponse
	fullPath := fmt.Sprintf("%s/api/%s", c.URI, path)
	c.logger.Debug(
		"Fetching http",
		"path", fullPath,
	)
	resp, err := c.makeRequest("GET", fullPath, nil, nil)
	if err != nil {
		c.logger.Error(
			logMsgErrAPICall,
			"endpoint", fullPath,
			"err", err.Error(),
		)
		return nil, err
	}
	response.NodeId = resp.Header.Get("x-artifactory-node-id")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			c.logger.Error(
				logMsgErrUnmarshall,
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
			"status", http.StatusNotFound,
		)
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
			status:   http.StatusNotFound,
		}
	}

	if !slices.Contains(httpSuccCodes, resp.StatusCode) {
		return c.procRespErr(resp, fullPath)
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
	resp, err := c.makeRequest("POST", fullPath, query, nil)
	if err != nil {
		c.logger.Error(
			logMsgErrAPICall,
			"endpoint", fullPath,
			"err", err.Error(),
		)
		return nil, err
	}
	response.NodeId = resp.Header.Get("x-artifactory-node-id")
	defer resp.Body.Close()
	if !slices.Contains(httpSuccCodes, resp.StatusCode) {
		return c.procRespErr(resp, fullPath)
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

// PostHTTP is a wrapper function for making all Post API calls
// Note: the API endpoint (e.g. "/artifactory" or "/access") needs to be part of path
func (c *Client) PostHTTP(path string, body []byte, headers *map[string]string) (*ApiResponse, error) {
	var response ApiResponse

	artifactoryURI := strings.TrimSuffix(c.URI, "/artifactory")
	fullPath := fmt.Sprintf("%s/%s", artifactoryURI, path)

	c.logger.Debug(
		"Posting http",
		"path", fullPath,
	)

	resp, err := c.makeRequest("POST", fullPath, body, &headers)
	c.logger.Debug(
		"Received response with",
		"status", resp.StatusCode,
	)
	if err != nil {
		c.logger.Error(
			logMsgErrAPICall,
			"endpoint", fullPath,
			"err", err.Error(),
		)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		var apiErrors APIErrors
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			c.logger.Error(
				logMsgErrUnmarshall,
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
			"status", http.StatusNotFound,
		)
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
			status:   http.StatusNotFound,
		}
	}

	if !slices.Contains(httpSuccCodes, resp.StatusCode) {
		return c.procRespErr(resp, fullPath)
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
