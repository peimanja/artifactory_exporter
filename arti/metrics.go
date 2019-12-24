package arti

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "arti"

var (
	Up = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "up",
			Help:      "State of Artifactory",
		})

	UserCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "security_users",
			Help:      "Number of artifactory users",
		},
		[]string{"realm"},
	)

	ArtifactCountTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "artifacts_count_total",
			Help:      "Total artifacts count stored in Artifactory",
		})
	ArtifactsSizeTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "artifacts_size_bytes_total",
			Help:      "Total artifacts Size stored in Artifactory in bytes",
		})
	BinariesCountTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "binaries_count_total",
			Help:      "Total binaries count stored in Artifactory",
		})
	BinariesSizeTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "binaries_size_bytes_total",
			Help:      "Total binaries Size stored in Artifactory in bytes",
		})
	FileStore = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "filestore_bytes",
			Help:      "Total space in the file store in bytes",
		},
		[]string{"storageType", "storageDirectory"},
	)
	FileStoreUsed = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "filestore_used_bytes",
			Help:      "Space used in the file store in bytes",
		},
		[]string{"storageType", "storageDirectory"},
	)
	FileStoreFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "filestore_free_bytes",
			Help:      "Space free in the file store in bytes",
		},
		[]string{"storageType", "storageDirectory"},
	)
	RepoUsedSpace = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "repo_used_bytes",
			Help:      "Space used by an Artifactory repository in bytes",
		},
		[]string{"name", "type", "package_type"},
	)
	RepoFolderCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "repo_folder_count",
			Help:      "Number of folders in an Artifactory repository",
		},
		[]string{"name", "type", "package_type"},
	)
	RepoFilesCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "repo_files_count",
			Help:      "Number files in an Artifactory repository",
		},
		[]string{"name", "type", "package_type"},
	)
	RepoItemsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "repo_items_count",
			Help:      "Number Items in an Artifactory repository",
		},
		[]string{"name", "type", "package_type"},
	)
	RepoPercentage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "repo_percentage",
			Help:      "Percentage of space used by an Artifactory repository",
		},
		[]string{"name", "type", "package_type"},
	)
	RepoCreatedArtifacts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "artifacts_created",
			Help:      "Created artifacts in an Artifactory repository",
		},
		[]string{"name", "type", "package_type", "minutes_ago"},
	)
	RepoDownloadedArtifacts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "artifacts_downloaded",
			Help:      "Downloaded artifacts from an Artifactory repository",
		},
		[]string{"name", "type", "package_type", "minutes_ago"},
	)
)

func init() {
	prometheus.MustRegister(Up)
	prometheus.MustRegister(UserCount)
	prometheus.MustRegister(ArtifactCountTotal)
	prometheus.MustRegister(ArtifactsSizeTotal)
	prometheus.MustRegister(BinariesCountTotal)
	prometheus.MustRegister(BinariesSizeTotal)
	prometheus.MustRegister(FileStore)
	prometheus.MustRegister(FileStoreUsed)
	prometheus.MustRegister(FileStoreFree)
	prometheus.MustRegister(RepoUsedSpace)
	prometheus.MustRegister(RepoFolderCount)
	prometheus.MustRegister(RepoFilesCount)
	prometheus.MustRegister(RepoItemsCount)
	prometheus.MustRegister(RepoPercentage)
	prometheus.MustRegister(RepoCreatedArtifacts)
	prometheus.MustRegister(RepoDownloadedArtifacts)
}
