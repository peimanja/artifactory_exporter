package artifactory

import (
	"encoding/json"
	"fmt"

	"github.com/go-kit/kit/log/level"
)

const federationMirrorsLagEndpoint = "federation/status/mirrorsLag"
const federationUnavailableMirrorsEndpoint = "federation/status/unavailableMirrors"
const federationRepoStatusEndpoint = "federation/status/repo"

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

// FederatedRepoStatus represents single element of API respond from federation/status/repo/<repoKey> endpoint
// We don't need all the fields but we'll leave them here for future use
type FederatedRepoStatus struct {
	LocalKey          string `json:"localKey"`
	BinariesTasksInfo struct {
		InProgressTasks int `json:"inProgressTasks"`
		FailingTasks    int `json:"failingTasks"`
	} `json:"binariesTasksInfo"`
	MirrorEventsStatusInfo []struct {
		RemoteUrl     string `json:"remoteUrl"`
		RemoteRepoKey string `json:"remoteRepoKey"`
		Status        string `json:"status"`
		CreateEvents  int    `json:"createEvents"`
		UpdateEvents  int    `json:"updateEvents"`
		DeleteEvents  int    `json:"deleteEvents"`
		PropsEvents   int    `json:"propsEvents"`
		ErrorEvents   int    `json:"errorEvents"`
		LagInMS       int    `json:"lagInMS"`
	} `json:"mirrorEventsStatusInfo"`
	FederatedArtifactStatus struct {
		CountFullyReplicateArtifacts         int `json:"countFullyReplicateArtifacts"`
		CountArtificiallyReplicatedArtifacts int `json:"countArtificiallyReplicatedArtifacts"`
	} `json:"federatedArtifactStatus"`
	NodeId string
}

// FetchFederatedRepoStatus makes the API call to federation/status/repo/<repoKey> endpoint and returns FederatedRepoStatus
func (c *Client) FetchFederatedRepoStatus(repoKey string) (FederatedRepoStatus, error) {
	var federatedRepoStatus FederatedRepoStatus
	level.Debug(c.logger).Log("msg", "Fetching federated repo status")
	resp, err := c.FetchHTTP(fmt.Sprintf("%s/%s", federationRepoStatusEndpoint, repoKey))
	if err != nil {
		if err.(*APIError).status == 404 {
			return federatedRepoStatus, nil
		}
		return federatedRepoStatus, err
	}

	federatedRepoStatus.NodeId = resp.NodeId

	if err := json.Unmarshal(resp.Body, &federatedRepoStatus); err != nil {
		level.Error(c.logger).Log("msg", "There was an issue when try to unmarshal federated repo status respond")
		return federatedRepoStatus, &UnmarshalError{
			message:  err.Error(),
			endpoint: fmt.Sprintf("%s/%s", federationRepoStatusEndpoint, repoKey),
		}
	}

	return federatedRepoStatus, nil
}
