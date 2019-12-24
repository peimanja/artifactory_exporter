package arti

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type APIClientConfig struct {
	Url  string
	User string
	Pass string
}

type SearchFeilds struct {
	Action string
	From   string
	To     string
	Repo   string
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
	log.WithFields(log.Fields{
		"method": "GET",
		"uri":    config.Url + "/api/" + path,
	}).Debugln("Making API request")

	u, err := url.Parse(config.Url)
	if err != nil {
		log.Fatal(err)
	}

	_, err = net.LookupHost(u.Hostname())
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", config.Url+"/api/"+path, nil)
	CheckErr(err)
	req.SetBasicAuth(config.User, config.Pass)
	resp, err := client.Do(req)
	CheckErr(err)
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	CheckErr(err)

	if resp.StatusCode == http.StatusOK {
		return bodyBytes, nil
	} else if resp.StatusCode == 401 {
		log.WithFields(log.Fields{
			"method":      "GET",
			"uri":         config.Url + "/api/" + path,
			"status_code": resp.StatusCode,
			"error":       string(bodyBytes),
		}).Fatalf("User `%s` is not authorized to access this endpoint", config.User)
	} else if resp.StatusCode == 404 {
		log.WithFields(log.Fields{
			"method":      "GET",
			"uri":         config.Url + "/api/" + path,
			"status_code": resp.StatusCode,
			"error":       string(bodyBytes),
		}).Fatalf("Looks like `%s` URL is wrong", config.Url)
	}

	return bodyBytes, errors.New(strconv.Itoa(resp.StatusCode))
}

func GetUp(config *APIClientConfig) float64 {
	path := "system/ping"
	resp, err := QueryArtiApi(config, path)

	bodyString := string(resp)
	if bodyString == "OK" {
		return 1
	}

	log.WithFields(log.Fields{
		"method":      "GET",
		"uri":         config.Url + "/api/" + path,
		"status_code": err,
		"error":       string(resp),
	}).Warn("There is an issue with Artifactory")

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
	resp, err := QueryArtiApi(config, "search/dates?dateFields="+s.Action+"&from="+s.From+"&to="+s.To+"&repos="+s.Repo)
	if err != nil {
		return 0
	}

	err = json.Unmarshal(resp, &results)
	if err != nil {
		return 0
	}

	return float64(len(results.Results))
}
