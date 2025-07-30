package artifactory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"slices"
	"strings"
	"sync"
)

const (
	logMsgErrAPICall    = "There was an error making API call"
	logMsgErrUnmarshall = "There was an error when trying to unmarshal the API Error"
	logMsgErrRespBody   = "There was an error reading response body"
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

func (c *Client) handleResponse(resp *http.Response, fullPath string) (*ApiResponse, error) {
	var apiErrors APIErrors
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error(
			logMsgErrRespBody,
			"err", err,
		)
		return nil, err
	}
	if !slices.Contains(httpSuccCodes, resp.StatusCode) {
		if err := json.Unmarshal(bodyBytes, &apiErrors); err != nil {
			c.logger.Error(
				logMsgErrUnmarshall,
				"err", err.Error(),
			)
			return nil, &UnmarshalError{
				message:  err.Error(),
				endpoint: fullPath,
			}
		}
		if resp.StatusCode == http.StatusNotFound {
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
		c.logger.Error(
			logMsgErrAPICall,
			"endpoint", fullPath,
			"err", fmt.Sprintf("%v", apiErrors.Errors),
			"status", resp.StatusCode,
		)
		return nil, &APIError{
			message:  fmt.Sprintf("%v", apiErrors.Errors),
			endpoint: fullPath,
			// status:   resp.StatusCode, // Maybe it would be worth returning it too? As with http.StatusNotFound.
		}
	}

	response := &ApiResponse{
		Body:   bodyBytes,
		NodeId: resp.Header.Get("x-artifactory-node-id"),
	}
	return response, nil
}

func (c *Client) makeCachedRequest(method string, path string, body []byte, headers **map[string]string) (*ApiResponse, error) {
	key := fmt.Sprintf("%s_%s_%s", method, path, body)
	cached := NewCached(key, c.responseCache, c.logger)

	var wg sync.WaitGroup
	wg.Add(2) // wait for sender and receiver goroutines
	defer wg.Done()
	go func() {
		wg.Wait()
		cached.Close()
	}()

	go func() {
		defer wg.Done()
		resp, err := c.makeRequest(method, path, body, headers)
		if err != nil {
			c.logger.Error(
				logMsgErrAPICall,
				"endpoint", path,
				"err", err.Error(),
			)
			cached.errors <- err
			return
		}
		defer resp.Body.Close()
		apiResp, err := c.handleResponse(resp, path)
		if err != nil {
			cached.errors <- err
			return
		}
		cached.responses <- apiResp
		cached.CacheResponse(apiResp)
	}()

	select {
	case err := <-cached.errors:
		c.logger.Warn(
			"Error while making request, fallback to cache",
			"method", method,
			"path", path,
			"err", err.Error(),
		)
		resp, exists := cached.GetCachedResponse()
		if exists {
			return resp, nil
		} else {
			return nil, err
		}
	case resp := <-cached.responses:
		return resp, nil
	}
}

// FetchHTTP is a wrapper function for making all Get API calls
func (c *Client) FetchHTTP(path string) (*ApiResponse, error) {
	fullPath := fmt.Sprintf("%s/api/%s", c.URI, path)
	c.logger.Debug(
		"Fetching http",
		"path", fullPath,
	)
	return c.makeCachedRequest("GET", fullPath, nil, nil)
}

// QueryAQL is a wrapper function for making an query to AQL endpoint
func (c *Client) QueryAQL(query []byte) (*ApiResponse, error) {
	fullPath := fmt.Sprintf("%s/api/search/aql", c.URI)
	c.logger.Debug(
		"Running AQL query",
		"path", fullPath,
	)
	return c.makeCachedRequest("POST", fullPath, query, nil)
}

// PostHTTP is a wrapper function for making all Post API calls
// Note: the API endpoint (e.g. "/artifactory" or "/access") needs to be part of path
func (c *Client) PostHTTP(path string, body []byte, headers *map[string]string) (*ApiResponse, error) {
	artifactoryURI := strings.TrimSuffix(c.URI, "/artifactory")
	fullPath := fmt.Sprintf("%s/%s", artifactoryURI, path)
	c.logger.Debug(
		"Posting http",
		"path", fullPath,
	)
	return c.makeCachedRequest("POST", fullPath, body, &headers)
}
