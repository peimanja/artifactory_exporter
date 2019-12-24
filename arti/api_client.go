package arti

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type APIClientConfig struct {
	url  string
	user string
	pass string
}

type SearchFeilds struct {
	action string
	from   string
	to     string
	repo   string
}

type ArtifactsAtr struct {
	Uri     string `json:"uri"`
	Created string `json:"created"`
}

type SearchResults struct {
	Results []ArtifactsAtr `json:"results"`
}

type User struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}

type BinariesSummary struct {
	BinariesCount  string `json:"binariesCount"`
	BinariesSize   string `json:"binariesSize"`
	ArtifactsSize  string `json:"artifactsSize"`
	Optimization   string `json:"optimization"`
	ItemsCount     string `json:"itemsCount"`
	ArtifactsCount string `json:"artifactsCount"`
}

type FileStoreSummary struct {
	StorageType      string `json:"storageType"`
	StorageDirectory string `json:"storageDirectory"`
	TotalSpace       string `json:"totalSpace"`
	UsedSpace        string `json:"usedSpace"`
	FreeSpace        string `json:"freeSpace"`
}

type RepositorySummary struct {
	Name         string `json:"repoKey"`
	Type         string `json:"repoType"`
	FoldersCount int    `json:"foldersCount"`
	FilesCount   int    `json:"filesCount"`
	UsedSpace    string `json:"usedSpace"`
	ItemsCount   int    `json:"itemsCount"`
	PackageType  string `json:"packageType"`
	Percentage   string `json:"percentage"`
}

type StorageSummary struct {
	BinariesSummary     BinariesSummary     `json:"binariesSummary"`
	FileStoreSummary    FileStoreSummary    `json:"fileStoreSummary"`
	RepositoriesSummary []RepositorySummary `json:"repositoriesSummaryList"`
}

type StorageInfo struct {
	Summary StorageSummary `json:"storageSummary"`
}

func QueryArtiApi(config *APIClientConfig, path string) ([]byte, error) {
	bodyBytes := []byte("")
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.url+"/api/"+path, nil)
	CheckErr(err)
	req.SetBasicAuth(config.user, config.pass)
	resp, err := client.Do(req)
	CheckErr(err)
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		CheckErr(err)
		return bodyBytes, err
	}
	return bodyBytes, err
}

func GetUp(config *APIClientConfig) float64 {
	resp, err := QueryArtiApi(config, "system/ping")
	CheckErr(err)
	bodyString := string(resp)
	if bodyString == "OK" {
		return 1
	}
	return 0
}

func GetUsers(config *APIClientConfig) ([]User, error) {
	var users []User
	resp, err := QueryArtiApi(config, "security/users")
	CheckErr(err)
	err = json.Unmarshal(resp, &users)
	return users, err
}

func GetStorageInfo(config *APIClientConfig) (StorageInfo, error) {
	var storageInfo StorageInfo
	resp, err := QueryArtiApi(config, "storageinfo")
	CheckErr(err)
	err = json.Unmarshal(resp, &storageInfo)
	return storageInfo, err
}

func GetRepoArtifacts(config *APIClientConfig, s SearchFeilds) float64 {
	var results SearchResults
	resp, err := QueryArtiApi(config, "search/dates?dateFields="+s.action+"&from="+s.from+"&to="+s.to+"&repos="+s.repo)
	if err != nil {
		return 0
	}

	err = json.Unmarshal(resp, &results)
	if err != nil {
		return 0
	}

	return float64(len(results.Results))
}
