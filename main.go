package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/peimanja/artifactory_exporter/arti"
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
	start := time.Now()

	var (
		listenAddress      = kingpin.Flag("web.listen-address", "Address to listen on for web interface and telemetry.").Envar("WEB_LISTEN_ADDR").Default(":9531").String()
		metricsPath        = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Envar("WEB_TELEMETRY_PATH").Default("/metrics").String()
		artiUser           = kingpin.Flag("artifactory.user", "User to access Artifactory.").Envar("ARTI_USER").Required().String()
		artiPassword       = kingpin.Flag("artifactory.password", "Password of the user accessing the Artifactory.").Envar("ARTI_PASSWORD").Required().String()
		artiScrapeURI      = kingpin.Flag("artifactory.scrape-uri", "URI on which to scrape Artifactory.").Envar("ARTI_SCRAPE_URI").Default("http://localhost:8081/artifactory").String()
		artiScrapeInterval = kingpin.Flag("artifactory.scrape-interval", "How often to scrape Artifactory in secoonds.").Envar("ARTI_SCRAPE_INTERVAL").Default("30").Int64()
		logLevel           = kingpin.Flag("exporter.debug", "Enable debug mode.").Envar("DEBUG").Default("false").Bool()
		ready              = false
	)

	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	config := &arti.APIClientConfig{
		*artiScrapeURI,
		*artiUser,
		*artiPassword,
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	if *logLevel {
		log.SetLevel(log.DebugLevel)
	}

	go func() {
		for {
			log.Infoln("Gathering metrics from Artifactory.")
			metrics := arti.Collect(config)
			arti.Updater(*metrics)

			log.Infof("Completed gathering metrics from Artifactory. Will update in: %ds", *artiScrapeInterval)
			ready = true
			time.Sleep(time.Duration(*artiScrapeInterval) * time.Second)
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

	for !ready {
		log.Debugln("Waiting for exporter to collect all metrics before starting the server")
	}

	elapsed := time.Since(start)
	log.Infof("Starting the server and listening on: %s", *listenAddress)
	log.Infof("Exposing metrics at: %s", *metricsPath)
	log.Infof("Initialization took: %s", elapsed)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
