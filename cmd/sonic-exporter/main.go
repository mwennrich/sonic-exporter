package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/mwennrich/sonic-exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/common/promslog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

func main() {
	var (
		webConfig   = webflag.AddFlags(kingpin.CommandLine, ":9101")
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)

	promslogConfig := &promslog.Config{}
	flag.AddFlags(kingpin.CommandLine, promslogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.Parse()

	logger := promslog.New(promslogConfig)

	interfaceCollector := collector.NewInterfaceCollector(logger)
	hwCollector := collector.NewHwCollector(logger)
	crmCollector := collector.NewCrmCollector(logger)
	prometheus.MustRegister(interfaceCollector)
	prometheus.MustRegister(hwCollector)
	prometheus.MustRegister(crmCollector)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
             <head><title>Sonic Exporter</title></head>
             <body>
             <h1>Sonic Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
		if err != nil {
			logger.ErrorContext(r.Context(), "Error writing response", "err", err)
		}
	})
	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := web.ListenAndServe(srv, webConfig, slog.Default()); err != nil {
		logger.ErrorContext(context.Background(), "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
