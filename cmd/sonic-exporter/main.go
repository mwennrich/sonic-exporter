package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/mwennrich/sonic-exporter/internal/collector"
)

func main() {
	var (
		webConfig   = webflag.AddFlags(kingpin.CommandLine, ":9101")
		metricsPath = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.HelpFlag.Short('h')
	kingpin.CommandLine.UsageWriter(os.Stdout)
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

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
			level.Error(logger).Log("msg", "Error writing response")
		}
	})
	srv := &http.Server{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	if err := web.ListenAndServe(srv, webConfig, slog.Default()); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
