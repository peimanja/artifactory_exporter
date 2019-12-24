package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"

	log "github.com/sirupsen/logrus"
)

var (
	histogramVec = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "prom_request_time",
		Help: "Time it has taken to retrieve the metrics",
	}, []string{"time"})
)

func newHandlerWithHistogram(handler http.Handler, histogram *prometheus.HistogramVec) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		status := http.StatusOK

		defer func() {
			histogram.WithLabelValues(fmt.Sprintf("%d", status)).Observe(time.Since(start).Seconds())
		}()

		if req.Method == http.MethodGet {
			handler.ServeHTTP(w, req)
			return
		}
		status = http.StatusBadRequest

		w.WriteHeader(status)
	})
}

func main() {
	var (
		listenAddress      = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Default(":9531").String()
		metricsPath        = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		artiUser           = kingpin.Flag("artifactory.user", "User to access Artifactory.").Envar("ARTI_USER").Required().String()
		artiPassword       = kingpin.Flag("artifactory.password", "Password of the user accessing the Artifactory.").Envar("ARTI_PASSWORD").Required().String()
		artiScrapeURI      = kingpin.Flag("artifactory.scrape-uri", "URI on which to scrape Artifactory.").Default("http://localhost:8081/artifactory").String()
		artiScrapeInterval = kingpin.Flag("artifactory.scrape-interval", "How often to scrape Artifactory in secoonds.").Default("30").Int64()
	)

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	config := &APIClientConfig{
		*artiScrapeURI,
		*artiUser,
		*artiPassword,
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	go func() {
		for {
			up.Set(GetUp(config))

			users, _ := GetUsers(config)
			countUsers := CountUsers(users)
			for _, count := range countUsers {
				userCount.WithLabelValues(count.Realm).Set(count.Count)
			}
			storageInfo, err := (GetStorageInfo(config))
			CheckErr(err)

			artifactCountTotal.Set(RemoveCommas(storageInfo.Summary.BinariesSummary.ArtifactsCount))

			binariesCountTotal.Set(RemoveCommas(storageInfo.Summary.BinariesSummary.BinariesCount))

			artifacts_size, err := BytesConverter(storageInfo.Summary.BinariesSummary.ArtifactsSize)
			CheckErr(err)
			artifactsSizeTotal.Set(artifacts_size)

			binaries_size, err := BytesConverter(storageInfo.Summary.BinariesSummary.BinariesSize)
			CheckErr(err)
			binariesSizeTotal.Set(binaries_size)

			fileStoreType := strings.ToLower(storageInfo.Summary.FileStoreSummary.StorageType)
			fileStoreDir := storageInfo.Summary.FileStoreSummary.StorageDirectory
			fileStoreSize, err := BytesConverter(storageInfo.Summary.FileStoreSummary.TotalSpace)
			CheckErr(err)
			fileStore.WithLabelValues(fileStoreType, fileStoreDir).Set(fileStoreSize)

			fileStoreUsedBytes, err := BytesConverter(storageInfo.Summary.FileStoreSummary.UsedSpace)
			CheckErr(err)
			fileStoreUsed.WithLabelValues(fileStoreType, fileStoreDir).Set(fileStoreUsedBytes)

			fileStoreFreeBytes, err := BytesConverter(storageInfo.Summary.FileStoreSummary.FreeSpace)
			CheckErr(err)
			fileStoreFree.WithLabelValues(fileStoreType, fileStoreDir).Set(fileStoreFreeBytes)

			searchArtifacts := &SearchFeilds{}
			searchArtifacts.from = strconv.FormatInt(FromEpoch(0), 10)
			for _, repo := range storageInfo.Summary.RepositoriesSummary {
				name := repo.Name
				if name == "TOTAL" {
					continue
				}
				repoType := strings.ToLower(repo.Type)
				foldersCount := float64(repo.FoldersCount)
				filesCount := float64(repo.FilesCount)
				itemsCount := float64(repo.ItemsCount)
				usedSpace, err := BytesConverter(repo.UsedSpace)
				CheckErr(err)
				packageType := strings.ToLower(repo.PackageType)
				percentage := RemoveCommas(repo.Percentage)

				repoUsedSpace.WithLabelValues(name, repoType, packageType).Set(usedSpace)
				repoFolderCount.WithLabelValues(name, repoType, packageType).Set(foldersCount)
				repoFilesCount.WithLabelValues(name, repoType, packageType).Set(filesCount)
				repoItemsCount.WithLabelValues(name, repoType, packageType).Set(itemsCount)
				repoPercentage.WithLabelValues(name, repoType, packageType).Set(percentage)

				searchArtifacts.repo = name
				for _, searchMinutesInterval := range []time.Duration{1, 5, 60} {
					searchArtifacts.from = strconv.FormatInt(fromEpoch(searchMinutesInterval), 10)
					for _, action := range []string{"created", "lastDownloaded"} {
						searchArtifacts.action = action
						numArtifacts := GetRepoArtifacts(config, *searchArtifacts)
						if action == "created" {
							repoCreatedArtifacts.WithLabelValues(name, repoType, packageType, strconv.FormatInt(int64(searchMinutesInterval), 10)).Set(numArtifacts)
						} else {
							repoDownloadedArtifacts.WithLabelValues(name, repoType, packageType, strconv.FormatInt(int64(searchMinutesInterval), 10)).Set(numArtifacts)
						}
					}
				}
			}

			log.Infoln("Finished gathering metrics from Artifactory. Will update in: " + strconv.FormatInt(*artiScrapeInterval, 10) + "s")
			time.Sleep(time.Duration(*artiScrapeInterval))
		}
	}()

	http.Handle(*metricsPath, newHandlerWithHistogram(promhttp.Handler(), histogramVec))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			 <head><title>Artifactory Exporter</title></head>
			 <body>
			 <h1>Artifactory Exporter</h1>
			 <p><a href='` + *metricsPath + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})
	log.Infof("Listening on: %s", *listenAddress)
	log.Infof("Exposing metrics at: %s", *metricsPath)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
