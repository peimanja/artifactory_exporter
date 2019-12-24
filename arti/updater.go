package arti

import (
	"runtime"
	"sync"
	"time"

	"github.com/prometheus/common/log"
)

func SetUserCount(countUsers <-chan UsersCount, wg *sync.WaitGroup) {
	for {
		count, ok := <-countUsers
		if ok {
			UserCount.WithLabelValues(count.Realm).Set(count.Count)
		} else {
			wg.Done()
			break
		}
	}
}

func SetRepoSummary(repos <-chan RepoSummary, wg *sync.WaitGroup) {
	for {
		repo, ok := <-repos
		if ok {
			RepoUsedSpace.WithLabelValues(repo.Name, repo.Type, repo.PackageType).Set(repo.UsedSpace)
			RepoFolderCount.WithLabelValues(repo.Name, repo.Type, repo.PackageType).Set(repo.FoldersCount)
			RepoFilesCount.WithLabelValues(repo.Name, repo.Type, repo.PackageType).Set(repo.FilesCount)
			RepoItemsCount.WithLabelValues(repo.Name, repo.Type, repo.PackageType).Set(repo.ItemsCount)
			RepoPercentage.WithLabelValues(repo.Name, repo.Type, repo.PackageType).Set(repo.Percentage)
		} else {
			wg.Done()
			break
		}
	}
}

func Updater(metrics MetricValues) {
	log.Infoln("Starting to updating the metrics")
	start := time.Now()

	Up.Set(metrics.ArtiUp)

	if metrics.ArtiUp == 1 {
		userCountChan := make(chan UsersCount, len(metrics.Users))
		var wg sync.WaitGroup

		for w := 1; w <= runtime.NumCPU(); w++ {
			go SetUserCount(userCountChan, &wg)
			wg.Add(1)
		}

		for _, v := range metrics.CountUsers {
			userCountChan <- v
		}

		close(userCountChan)
		wg.Wait()

		ArtifactCountTotal.Set(metrics.ArtifactCountTotal)

		BinariesCountTotal.Set(metrics.BinariesCountTotal)

		ArtifactsSizeTotal.Set(metrics.ArtifactsSize)

		BinariesSizeTotal.Set(metrics.BinariesSize)

		FileStore.WithLabelValues(metrics.FileStoreType, metrics.FileStoreDir).Set(metrics.FileStoreSize)

		FileStoreUsed.WithLabelValues(metrics.FileStoreType, metrics.FileStoreDir).Set(metrics.FileStoreUsedBytes)

		FileStoreFree.WithLabelValues(metrics.FileStoreType, metrics.FileStoreDir).Set(metrics.FileStoreFreeBytes)

		repoSumChan := make(chan RepoSummary, len(metrics.RepositoriesSummary))
		for w := 1; w <= runtime.NumCPU(); w++ {
			go SetRepoSummary(repoSumChan, &wg)
			wg.Add(1)
		}

		for _, r := range metrics.RepositoriesSummary {
			repoSumChan <- r
		}

		close(repoSumChan)
		wg.Wait()

		elapsed := time.Since(start)
		log.Infof("Updated metrics in: %s", elapsed)
	}
}
