package artifactory

import (
	"encoding/json"
	"errors"
	"fmt"
)

const replicationEndpoint = "replications"
const replicationStatusEndpoint = "replication"

// Replication represents single element of API respond from replication endpoint
type Replication struct {
	ReplicationType                 string `json:"replicationType"`
	Enabled                         bool   `json:"enabled"`
	CronExp                         string `json:"cronExp"`
	SyncDeletes                     bool   `json:"syncDeletes"`
	SyncProperties                  bool   `json:"syncProperties"`
	PathPrefix                      string `json:"pathPrefix"`
	RepoKey                         string `json:"repoKey"`
	URL                             string `json:"url"`
	EnableEventReplication          bool   `json:"enableEventReplication"`
	CheckBinaryExistenceInFilestore bool   `json:"checkBinaryExistenceInFilestore"`
	SyncStatistics                  bool   `json:"syncStatistics"`
	Status                          string `json:"status"`
}

type Replications struct {
	Replications []Replication
	NodeId       string
}

type ReplicationStatus struct {
	Status string `json:"status"`
}

// FetchReplications makes the API call to replication endpoint and returns []Replication
func (c *Client) FetchReplications() (Replications, error) {
	var replications Replications
	c.logger.Debug("Fetching replications stats")
	resp, err := c.FetchHTTP(replicationEndpoint)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.status == 404 {
			return replications, nil
		}
		return replications, err
	}
	replications.NodeId = resp.NodeId

	if err := json.Unmarshal(resp.Body, &replications.Replications); err != nil {
		c.logger.Error("There was an issue when try to unmarshal replication respond")
		return replications, &UnmarshalError{
			message:  err.Error(),
			endpoint: replicationEndpoint,
		}
	}

	if c.optionalMetrics.ReplicationStatus {
		c.logger.Debug("Fetching replications status")
		for i, replication := range replications.Replications {
			var status ReplicationStatus
			if replication.Enabled {
				statusResp, err := c.FetchHTTP(fmt.Sprintf("%s/%s", replicationStatusEndpoint, replication.RepoKey))
				if err != nil {
					return replications, err
				}
				if err := json.Unmarshal(statusResp.Body, &status); err != nil {
					c.logger.Error("There was an issue when try to unmarshal replication status respond")
					return replications, &UnmarshalError{
						message:  err.Error(),
						endpoint: fmt.Sprintf("%s/%s", replicationStatusEndpoint, replication.RepoKey),
					}
				}
				replications.Replications[i].Status = status.Status
			}
		}
	}

	return replications, nil
}
