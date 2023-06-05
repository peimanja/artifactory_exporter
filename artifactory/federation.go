package artifactory

import (
	"encoding/json"

	"github.com/go-kit/log/level"
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

// FetchMirrorLags makes the API call to federation/status/mirrorsLag endpoint and returns []MirrorLag
func (c *Client) FetchMirrorLags() (MirrorLags, error) {
	var mirrorLags MirrorLags
	level.Debug(c.logger).Log("msg", "Fetching mirror lags")
	resp, err := c.FetchHTTP(federationMirrorsLagEndpoint)
	if err != nil {
		if err.(*APIError).status == 404 {
			return mirrorLags, nil
		}
		return mirrorLags, err
	}
	mirrorLags.NodeId = resp.NodeId

	if err := json.Unmarshal(resp.Body, &mirrorLags.MirrorLags); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal mirror lags respond")
		return mirrorLags, &UnmarshalError{
			message:  err.Error(),
			endpoint: federationMirrorsLagEndpoint,
		}
	}

	return mirrorLags, nil
}

// UnavailableMirror represents single element of API respond from federation/status/unavailableMirrors endpoint
type UnavailableMirror struct {
	LocalRepoKey  string `json:"localRepoKey"`
	RemoteUrl     string `json:"remoteUrl"`
	RemoteRepoKey string `json:"remoteRepoKey"`
	Status        string `json:"status"`
}

type UnavailableMirrors struct {
	UnavailableMirrors []UnavailableMirror
	NodeId             string
}

// FetchUnavailableMirrors makes the API call to federation/status/unavailableMirrors endpoint and returns []UnavailableMirror
func (c *Client) FetchUnavailableMirrors() (UnavailableMirrors, error) {
	var unavailableMirrors UnavailableMirrors
	level.Debug(c.logger).Log("msg", "Fetching unavailable mirrors")
	resp, err := c.FetchHTTP(federationUnavailableMirrorsEndpoint)
	if err != nil {
		if err.(*APIError).status == 404 {
			return unavailableMirrors, nil
		}
		return unavailableMirrors, err
	}
	unavailableMirrors.NodeId = resp.NodeId

	if err := json.Unmarshal(resp.Body, &unavailableMirrors.UnavailableMirrors); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal unavailable mirrors respond")
		return unavailableMirrors, &UnmarshalError{
			message:  err.Error(),
			endpoint: federationUnavailableMirrorsEndpoint,
		}
	}

	return unavailableMirrors, nil
}
