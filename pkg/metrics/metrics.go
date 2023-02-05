package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"soko/pkg/app/utils"
	"soko/pkg/database"
	"soko/pkg/models"
	"time"
)

var (
	UpdateAges = map[string]prometheus.Gauge{
		"dependencies": promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "pgo_update_age",
			Help:        "The age of the last update",
			ConstLabels: prometheus.Labels{"type": "dependencies"},
		}),
		"pkgcheck": promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "pgo_update_age",
			Help:        "The age of the last update",
			ConstLabels: prometheus.Labels{"type": "pkgcheck"},
		}),
		"pullrequests": promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "pgo_update_age",
			Help:        "The age of the last update",
			ConstLabels: prometheus.Labels{"type": "pullrequests"},
		}),
		"bugs": promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "pgo_update_age",
			Help:        "The age of the last update",
			ConstLabels: prometheus.Labels{"type": "bugs"},
		}),
		"projects": promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "pgo_update_age",
			Help:        "The age of the last update",
			ConstLabels: prometheus.Labels{"type": "projects"},
		}),
		"maintainers": promauto.NewGauge(prometheus.GaugeOpts{
			Name:        "pgo_update_age",
			Help:        "The age of the last update",
			ConstLabels: prometheus.Labels{"type": "maintainers"},
		}),
	}

	LastCommitAge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "pgo_last_commit_age",
		Help: "The age of the last commit",
	})
)

func Update() {

	database.Connect()
	defer database.DBCon.Close()

	var applicationData []*models.Application
	database.DBCon.Model(&applicationData).Select()

	for _, applications := range applicationData {
		if metric, ok := UpdateAges[applications.Id]; ok {
			metric.Set(time.Since(applications.LastUpdate).Seconds())
		}
	}

	lastCommit := &models.Commit{Id: utils.GetApplicationData().LastCommit}
	err := database.DBCon.Select(lastCommit)
	if err == nil {
		LastCommitAge.Set(time.Since(lastCommit.CommitterDate).Seconds())
	}

}
