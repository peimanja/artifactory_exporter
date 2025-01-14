package artifactory

import (
	"encoding/json"
)

const (
	accessFederationValidateEndpoint = "access/api/v1/system/federation/validate_server"
)

type AccessFederationValid struct {
	Status bool
	NodeId string
}

// FetchAccessFederationValidStatus checks one of the federation endpoints to see if federation is enabled
func (c *Client) FetchAccessFederationValidStatus() (AccessFederationValid, error) {
	accessFederationValid := AccessFederationValid{Status: false}

	// Use ping endpoint to retrieve nodeID, since this is not returned by access API
	resp, err := c.FetchHTTP(pingEndpoint)
	if err != nil {
		return accessFederationValid, err
	}
	accessFederationValid.NodeId = resp.NodeId

	jsonBody := map[string]string{
		"url": c.accessFederationTarget,
	}
	jsonBytes, err := json.Marshal(jsonBody)
	if err != nil {
		c.logger.Error("issue when trying to marshal JSON body")
		return accessFederationValid, err
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	c.logger.Debug(
		"Fetching JFrog Access Federation validation status",
		"endpoint", accessFederationValidateEndpoint,
		"target", c.accessFederationTarget,
	)
	_, err = c.PostHTTP(accessFederationValidateEndpoint, jsonBytes, &headers)
	if err != nil {
		return accessFederationValid, err
	}
	accessFederationValid.Status = true
	return accessFederationValid, nil
}
