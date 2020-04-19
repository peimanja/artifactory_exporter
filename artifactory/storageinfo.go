package artifactory

import (
	"encoding/json"

	"github.com/go-kit/kit/log/level"
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
		ItemsCount     string `json:""`
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
}

// FetchStorageInfo makes the API call to storageinfo endpoint and returns StorageInfo
func (c *Client) FetchStorageInfo() (StorageInfo, error) {
	var storageInfo StorageInfo
	level.Debug(c.logger).Log("msg", "Fetching storage info stats")
	resp, err := c.fetchHTTP(storageInfoEndpoint)
	if err != nil {
		return storageInfo, err
	}
	if err := json.Unmarshal(resp, &storageInfo); err != nil {
		level.Error(c.logger).Log("err", "There was an issue getting storageInfo respond")
		return storageInfo, err
	}
	return storageInfo, nil
}
