package main

import (
	"net/http"
	"os"

	"github.com/go-kit/kit/log/level"
	"github.com/peimanja/artifactory_exporter/collector"
	"github.com/peimanja/artifactory_exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

func main() {

	c, err := config.NewConfig()
	if err != nil {
		log.Errorf("Error creating the config. err: %s", err)
		os.Exit(1)
	}

	exporter, err := collector.NewExporter(c.ArtiScrapeURI, *c.Credentials, c.AuthMethod, c.ArtiSSLVerify, c.ArtiTimeout, c.Logger)
	if err != nil {
		level.Error(c.Logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	prometheus.MustRegister(exporter)

	level.Info(c.Logger).Log("msg", "Listening on address", "address", c.ListenAddress)
	http.Handle(c.MetricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>JFrog Exporter</title></head>
             <body>
             <h1>JFrog Exporter</h1>
             <p><a href='` + c.MetricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	if err := http.ListenAndServe(c.ListenAddress, nil); err != nil {
		level.Error(c.Logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
