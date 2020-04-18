package collector

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/config"
)

func (e *Exporter) fetchHTTP(uri string, path string, cred config.Credentials, authMethod string, sslVerify bool, timeout time.Duration) ([]byte, error) {
	fullPath := fmt.Sprintf("%s/api/%s", uri, path)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	level.Debug(e.logger).Log("msg", "Fetching http", "path", fullPath)
	req, err := http.NewRequest("GET", fullPath, nil)
	if err != nil {
		return nil, err
	}
	if authMethod == "userPass" {
		req.SetBasicAuth(cred.Username, cred.Password)
	} else if authMethod == "accessToken" {
		req.Header.Add("Authorization", "Bearer "+cred.AccessToken)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
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
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !e.sslVerify}}
	client := http.Client{
		Timeout:   e.timeout,
		Transport: tr,
	}
	level.Debug(e.logger).Log("msg", "Running AQL query", "path", fullPath)
	req, err := http.NewRequest("POST", fullPath, bytes.NewBuffer(query))
	req.Header = http.Header{"Content-Type": {"text/plain"}}
	if err != nil {
		return nil, err
	}

	if e.authMethod == "userPass" {
		req.SetBasicAuth(e.cred.Username, e.cred.Password)
	} else if e.authMethod == "accessToken" {
		req.Header.Add("Authorization", "Bearer "+e.cred.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
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

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
