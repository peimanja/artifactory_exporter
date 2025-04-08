package artifactory

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"time"
)

const federationMirrorsLagEndpoint = "federation/status/mirrorsLag"
const federationUnavailableMirrorsEndpoint = "federation/status/unavailableMirrors"

// IsFederationEnabled checks one of the federation endpoints to see if federation is enabled
func (c *Client) IsFederationEnabled() bool {
	_, err := c.FetchHTTP(federationUnavailableMirrorsEndpoint)
	if err != nil {
		return false
	}
	return true
}

// MirrorLag represents single element of API respond from federation/status/mirrorsLag endpoint
type MirrorLag struct {
	LocalRepoKey  string `json:"localRepoKey"`
	RemoteUrl     string `json:"remoteUrl"`
	RemoteRepoKey string `json:"remoteRepoKey"`
	LagInMS       int    `json:"lagInMS"`
}

type MirrorLags struct {
	MirrorLags []MirrorLag
	NodeId     string
}

type UnavailableMirror struct {
	RepoKey       string `json:"repoKey"`
	NodeId        string `json:"nodeId"`
	Status        string `json:"status"`
	LocalRepoKey  string `json:"localRepoKey"`
	RemoteUrl     string `json:"remoteUrl"`
	RemoteRepoKey string `json:"remoteRepoKey"`
}

type UnavailableMirrors struct {
	UnavailableMirrors []UnavailableMirror
	NodeId             string
}

// FetchMirrorLags makes the API call to federation/status/mirrorsLag endpoint and returns []MirrorLag
func (c *Client) FetchMirrorLags() (MirrorLags, error) {
	var mirrorLags MirrorLags
	c.logger.Debug("Fetching mirror lags")

	resp, err := c.FetchHTTP(federationMirrorsLagEndpoint)
	if err != nil {
		var apiErr *APIError
		var urlErr *url.Error
		if errors.As(err, &apiErr) && apiErr.status == 404 {
			return mirrorLags, nil
		} else if errors.As(err, &urlErr) {
			c.logger.Error("URL error while fetching mirror lags: ", urlErr)
			return mirrorLags, err
		} else {
			return mirrorLags, err
		}
	}
	mirrorLags.NodeId = resp.NodeId

	if err := json.Unmarshal(resp.Body, &mirrorLags.MirrorLags); err != nil {
		c.logger.Error("There was an issue when trying to unmarshal mirror lags response: ", err)
		return mirrorLags, &UnmarshalError{
			message:  err.Error(),
			endpoint: federationMirrorsLagEndpoint,
		}
	}

	return mirrorLags, nil
}

// FetchUnavailableMirrors makes the API call to federation/status/unavailableMirrors endpoint and returns []UnavailableMirror
func (c *Client) FetchUnavailableMirrors() (UnavailableMirrors, error) {
	var unavailableMirrors UnavailableMirrors
	c.logger.Debug("Fetching unavailable mirrors")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.FetchHTTPWithContext(ctx, federationUnavailableMirrorsEndpoint)
	if err != nil {
		var apiErr *APIError
		var urlErr *url.Error
		if errors.As(err, &apiErr) && apiErr.status == 404 {
			return unavailableMirrors, nil
		} else if errors.As(err, &urlErr) {
			c.logger.Error("URL error while fetching unavailable mirrors: ", urlErr)
			return unavailableMirrors, err
		} else {
			return unavailableMirrors, err
		}
	}
	unavailableMirrors.NodeId = resp.NodeId

	if err := json.Unmarshal(resp.Body, &unavailableMirrors.UnavailableMirrors); err != nil {
		c.logger.Error("There was an issue when trying to unmarshal unavailable mirrors response: ", err)
		return unavailableMirrors, &UnmarshalError{
			message:  err.Error(),
			endpoint: federationUnavailableMirrorsEndpoint,
		}
	}

	return unavailableMirrors, nil
}
