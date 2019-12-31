package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/peimanja/artifactory_exporter/arti"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	start := time.Now()

	var (
		ready = false
	)

	config, err := arti.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if config.Debug {
		log.SetLevel(log.DebugLevel)
	}

	go func() {
		for {
			log.Infoln("Gathering metrics from Artifactory.")
			metrics := arti.Collect(config.ApiConfig)
			arti.Updater(*metrics)

			log.Infof("Completed gathering metrics from Artifactory. Will update in: %ds", config.ArtiScrapeInterval)
			ready = true
			time.Sleep(time.Duration(config.ArtiScrapeInterval) * time.Second)
		}
	}()

	http.Handle(config.MetricsPath, newHandlerWithHistogram(promhttp.Handler(), histogramVec))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			 <head><title>Artifactory Exporter</title></head>
			 <body>
			 <h1>Artifactory Exporter</h1>
			 <p><a href='` + config.MetricsPath + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})

	for !ready {
		log.Debugln("Waiting for exporter to collect all metrics before starting the server")
	}

	elapsed := time.Since(start)
	log.Infof("Starting the server and listening on: %s", config.ListenAddress)
	log.Infof("Exposing metrics at: %s", config.MetricsPath)
	log.Infof("Initialization took: %s", elapsed)
	log.Fatal(http.ListenAndServe(config.ListenAddress, nil))
}
