package artifactory

import (
	"encoding/json"

	"github.com/go-kit/kit/log/level"
)

const replicationEndpoint = "replications"

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
}

// FetchReplications makes the API call to replication endpoint and returns []Replication
func (c *Client) FetchReplications() ([]Replication, error) {
	var replications []Replication
	level.Debug(c.logger).Log("msg", "Fetching replications stats")
	resp, err := c.fetchHTTP(replicationEndpoint)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(resp, &replications); err != nil {
		level.Warn(c.logger).Log("msg", "There was an issue getting replication respond")
		return replications, err
	}
	return replications, nil
}
