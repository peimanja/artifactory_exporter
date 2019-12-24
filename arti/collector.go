package arti

import (
	"strings"

	log "github.com/sirupsen/logrus"
)

type MetricValues struct {
	ArtiUp              float64
	ArtifactsSize       float64
	BinariesSize        float64
	Users               []User
	CountUsers          []UsersCount
	StorageInfo         StorageInfo
	FileStoreType       string
	FileStoreDir        string
	FileStoreSize       float64
	FileStoreUsedBytes  float64
	FileStoreFreeBytes  float64
	ArtifactCountTotal  float64
	BinariesCountTotal  float64
	RepositoriesSummary []RepoSummary
}

type RepoSummary struct {
	Name         string
	Type         string
	FoldersCount float64
	FilesCount   float64
	UsedSpace    float64
	ItemsCount   float64
	PackageType  string
	Percentage   float64
}

func Collect(config *APIClientConfig) *MetricValues {
	metrics := &MetricValues{}

	log.Debugln("Collecting Artifactory `up` Metric")
	metrics.ArtiUp = GetUp(config)

	log.Debugln("Getting list of Artifactory users")
	users, err := GetUsers(config)
	CheckErr(err)

	log.Debugln("Counting Number of Internal and SAML users")
	metrics.CountUsers = CountUsers(users)

	log.Debugln("Collecting the Storage info")
	storageInfo, err := (GetStorageInfo(config))
	CheckErr(err)

	log.Debugln("Calculating Total Artifacts Size metrics")
	metrics.ArtifactsSize, err = BytesConverter(storageInfo.Summary.BinariesSummary.ArtifactsSize)
	CheckErr(err)

	log.Debugln("Calculating Total Binaries Size metrics")
	metrics.BinariesSize, err = BytesConverter(storageInfo.Summary.BinariesSummary.BinariesSize)
	CheckErr(err)

	log.Debugln("Collecting FileStore metrics")
	metrics.FileStoreType = strings.ToLower(storageInfo.Summary.FileStoreSummary.StorageType)
	metrics.FileStoreDir = storageInfo.Summary.FileStoreSummary.StorageDirectory

	metrics.FileStoreSize, err = BytesConverter(storageInfo.Summary.FileStoreSummary.TotalSpace)
	CheckErr(err)

	metrics.FileStoreUsedBytes, err = BytesConverter(storageInfo.Summary.FileStoreSummary.UsedSpace)
	CheckErr(err)

	metrics.FileStoreFreeBytes, err = BytesConverter(storageInfo.Summary.FileStoreSummary.FreeSpace)
	CheckErr(err)

	log.Debugln("Collecting Total Artifacts Count")
	metrics.ArtifactCountTotal = RemoveCommas(storageInfo.Summary.BinariesSummary.ArtifactsCount)

	log.Debugln("Collecting Total Binaries Count")
	metrics.BinariesCountTotal = RemoveCommas(storageInfo.Summary.BinariesSummary.BinariesCount)

	log.Debugln("Collecting Artifactory Repositories metrics")
	repoSummary := &RepoSummary{}
	for _, repo := range storageInfo.Summary.RepositoriesSummary {
		if repo.Name == "TOTAL" {
			continue
		}
		log.Debugf("Collecting %s Repository metrics", repo.Name)
		repoSummary.Name = repo.Name
		repoSummary.Type = strings.ToLower(repo.Type)
		repoSummary.FoldersCount = float64(repo.FoldersCount)
		repoSummary.FilesCount = float64(repo.FilesCount)
		repoSummary.ItemsCount = float64(repo.ItemsCount)
		repoSummary.PackageType = strings.ToLower(repo.PackageType)
		repoSummary.UsedSpace, err = BytesConverter(repo.UsedSpace)
		repoSummary.Percentage = RemoveCommas(repo.Percentage)
		CheckErr(err)

		metrics.RepositoriesSummary = append(metrics.RepositoriesSummary, *repoSummary)

	}

	return metrics
}
