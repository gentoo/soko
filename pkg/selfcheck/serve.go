package selfcheck

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"soko/pkg/config"
	"soko/pkg/logger"
	"soko/pkg/metrics"
)

// Serve is used to serve the web application
func Serve() {

	// prometheus metrics
	http.Handle("/metrics", metricsHandler())

	logger.Info.Println("Serving on port: " + config.Port())
	log.Fatal(http.ListenAndServe(":"+config.Port(), nil))

}

// metricsHandler is used as default middleware to update the metrics
func metricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.Update()
		promhttp.Handler().ServeHTTP(w, r)
	})
}
