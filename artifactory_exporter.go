package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	"github.com/peimanja/artifactory_exporter/collector"
	"github.com/peimanja/artifactory_exporter/config"
	"github.com/peimanja/artifactory_exporter/logger"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		logger.New(logger.EmptyConfig).Error(
			"Error creating the config.",
			"err", err.Error(),
		)
		os.Exit(1)
	}

	exporter, err := collector.NewExporter(conf)
	if err != nil {
		conf.Logger.Error(
			"Error creating an exporter",
			"err", err.Error(),
		)
		os.Exit(1)
	}
	collector.InitMetrics(exporter)
	prometheus.MustRegister(exporter)
	conf.Logger.Info(
		"Starting artifactory_exporter",
		"version", version.Info(),
	)
	conf.Logger.Info(
		"Build context",
		"context", version.BuildContext(),
	)
	conf.Logger.Info(
		"Listening on address",
		"address", conf.ListenAddress,
	)
	http.HandleFunc(conf.MetricsPath, func(w http.ResponseWriter, r *http.Request) {
		conf.Logger.Debug(
			"Prometheus scrape",
			"remote", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)
		promhttp.Handler().ServeHTTP(w, r)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>JFrog Artifactory Exporter</title></head>
             <body>
             <h1>JFrog Exporter</h1>
             <p><a href='` + conf.MetricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	http.HandleFunc("/-/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})
	http.HandleFunc("/-/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})
	if err := http.ListenAndServe(conf.ListenAddress, nil); err != nil {
		conf.Logger.Error(
			"Error starting HTTP server",
			"err", err.Error(),
		)
		os.Exit(1)
	}
}
