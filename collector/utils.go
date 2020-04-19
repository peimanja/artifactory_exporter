package collector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/kit/log/level"
)

type errorResponse struct {
	Errors []struct {
		Status  int    `json:"status,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"errors,omitempty"`
}

func (e *Exporter) fetchHTTP(path string) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/api/%s", e.URI, path)
	level.Debug(e.logger).Log("msg", "Fetching http", "path", fullPath)
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	switch e.authMethod {
	case "userPass":
		req.SetBasicAuth(e.cred.Username, e.cred.Password)
	case "accessToken":
		req.Header.Add("Authorization", "Bearer "+e.cred.AccessToken)
	default:
		return nil, fmt.Errorf("Artifactory Auth method is not supported")
	}
	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		e.totalAPIErrors.Inc()
		var apiError errorResponse
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &apiError)
		level.Error(e.logger).Log("msg", "There was an error making API call", "endpoint", fullPath, "err", apiError.Errors[0].Message, "status", apiError.Errors[0].Status)
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return bodyBytes, nil
}

func (e *Exporter) queryAQL(query []byte) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/api/search/aql", e.URI)
	level.Debug(e.logger).Log("msg", "Running AQL query", "path", fullPath)
	req, err := http.NewRequest("POST", fullPath, bytes.NewBuffer(query))
	req.Header = http.Header{"Content-Type": {"text/plain"}}
	if err != nil {
		return nil, err
	}
	switch e.authMethod {
	case "userPass":
		req.SetBasicAuth(e.cred.Username, e.cred.Password)
	case "accessToken":
		req.Header.Add("Authorization", "Bearer "+e.cred.AccessToken)
	default:
		return nil, fmt.Errorf("Artifactory Auth method is not supported")
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		e.totalAPIErrors.Inc()
		var apiError errorResponse
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &apiError)
		level.Error(e.logger).Log("msg", "There was an error making API call", "endpoint", fullPath, "err", apiError.Errors[0].Message, "status", apiError.Errors[0].Status)
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return bodyBytes, nil
}

func (e *Exporter) removeCommas(str string) (float64, error) {
	level.Debug(e.logger).Log("msg", "Removing other characters to extract number from string")
	reg, err := regexp.Compile("[^0-9.]+")
	if err != nil {
		return 0, err
	}
	strArray := strings.Fields(str)
	convertedStr, err := strconv.ParseFloat(reg.ReplaceAllString(strArray[0], ""), 64)
	if err != nil {
		return 0, err
	}
	level.Debug(e.logger).Log("msg", "Successfully converted string to number", "string", str, "number", convertedStr)
	return convertedStr, nil
}

func (e *Exporter) bytesConverter(str string) (float64, error) {
	var bytesValue float64
	level.Debug(e.logger).Log("msg", "Converting size to bytes")
	num, err := e.removeCommas(str)
	if err != nil {
		return 0, err
	}

	if strings.Contains(str, "bytes") {
		bytesValue = num
	} else if strings.Contains(str, "KB") {
		bytesValue = num * 1024
	} else if strings.Contains(str, "MB") {
		bytesValue = num * 1024 * 1024
	} else if strings.Contains(str, "GB") {
		bytesValue = num * 1024 * 1024 * 1024
	} else if strings.Contains(str, "TB") {
		bytesValue = num * 1024 * 1024 * 1024 * 1024
	} else {
		return 0, fmt.Errorf("Could not convert %s to bytes", str)
	}
	level.Debug(e.logger).Log("msg", "Successfully converted string to bytes", "string", str, "value", bytesValue)
	return bytesValue, nil
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
