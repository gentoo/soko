package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	MissingVersions = map[string]prometheus.Gauge {}
	MissingPackages = map[string]prometheus.Gauge {}
)

