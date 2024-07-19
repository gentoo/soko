// SPDX-License-Identifier: GPL-2.0-only
package selfcheck

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"soko/pkg/config"
	"soko/pkg/metrics"
)

// Serve is used to serve the web application
func Serve() {
	// prometheus metrics
	http.Handle("/metrics", metricsHandler())

	address := ":" + config.Port()
	slog.Info("Serving self-check", slog.String("address", address))
	err := http.ListenAndServe(address, nil)
	slog.Error("exited server", "err", err)
	os.Exit(1)
}

// metricsHandler is used as default middleware to update the metrics
func metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.Update()
		promhttp.Handler().ServeHTTP(w, r)
	})
}
