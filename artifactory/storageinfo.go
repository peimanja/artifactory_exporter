package artifactory

import (
	"encoding/json"
)

const (
	storageInfoEndpoint = "storageinfo"
)

// StorageInfo represents API respond from license storageinfo
type StorageInfo struct {
	BinariesSummary struct {
		BinariesCount  string `json:"binariesCount"`
		BinariesSize   string `json:"binariesSize"`
		ArtifactsSize  string `json:"artifactsSize"`
		Optimization   string `json:"optimization"`
		ItemsCount     string `json:"itemsCount"`
		ArtifactsCount string `json:"artifactsCount"`
	} `json:"binariesSummary"`
	FileStoreSummary struct {
		StorageType      string `json:"storageType"`
		StorageDirectory string `json:"storageDirectory"`
		TotalSpace       string `json:"totalSpace"`
		UsedSpace        string `json:"usedSpace"`
		FreeSpace        string `json:"freeSpace"`
	} `json:"fileStoreSummary"`
	RepositoriesSummaryList []struct {
		RepoKey      string `json:"repoKey"`
		RepoType     string `json:"repoType"`
		FoldersCount int    `json:"foldersCount"`
		FilesCount   int    `json:"filesCount"`
		UsedSpace    string `json:"usedSpace"`
		ItemsCount   int    `json:"itemsCount"`
		PackageType  string `json:"packageType"`
		Percentage   string `json:"percentage"`
	} `json:"repositoriesSummaryList"`
	NodeId string
}

// FetchStorageInfo makes the API call to storageinfo endpoint and returns StorageInfo
func (c *Client) FetchStorageInfo() (StorageInfo, error) {
	var storageInfo StorageInfo
	c.logger.Debug("Fetching storage info stats")
	resp, err := c.FetchHTTP(storageInfoEndpoint)
	if err != nil {
		return storageInfo, err
	}
	storageInfo.NodeId = resp.NodeId
	if err := json.Unmarshal(resp.Body, &storageInfo); err != nil {
		c.logger.Error("There was an issue when try to unmarshal storageInfo respond")
		return storageInfo, &UnmarshalError{
			message:  err.Error(),
			endpoint: storageInfoEndpoint,
		}
	}
	return storageInfo, nil
}
