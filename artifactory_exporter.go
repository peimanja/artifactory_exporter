package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	"github.com/peimanja/artifactory_exporter/collector"
	"github.com/peimanja/artifactory_exporter/config"
)

func main() {
	conf, err := config.NewConfig()
	if err != nil {
		slog.Error(
			"Error creating the config.",
			"err", err.Error(),
		)
		os.Exit(1)
	}

	exporter, err := collector.NewExporter(conf)
	if err != nil {
		level.Error(conf.Logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)
	level.Info(conf.Logger).Log("msg", "Starting artifactory_exporter", "version", version.Info())
	level.Info(conf.Logger).Log("msg", "Build context", "context", version.BuildContext())
	level.Info(conf.Logger).Log("msg", "Listening on address", "address", conf.ListenAddress)
	http.Handle(conf.MetricsPath, promhttp.Handler())
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
		level.Error(conf.Logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
