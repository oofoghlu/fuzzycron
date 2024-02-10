package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	CronCreation = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cron_create_total",
			Help: "Number of CronJob creations",
		}, []string{"cron_job_name", "namespace"},
	)
	CronCreationError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cron_create_error_total",
			Help: "Number of CronJob creation errors",
		}, []string{"cron_job_name", "namespace"},
	)
	CronUpdate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cron_update_total",
			Help: "Number of CronJob updates",
		}, []string{"cron_job_name", "namespace"},
	)
	CronUpdateError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cron_update_error_total",
			Help: "Number of CronJob update errors",
		}, []string{"cron_job_name", "namespace"},
	)
	FuzzyStatusUpdate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fuzzy_cron_status_update_total",
			Help: "Number of FuzzyCronJob status updates",
		}, []string{"cron_job_name", "namespace"},
	)
	FuzzyStatusUpdateError = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fuzzy_cron_status_update_error_total",
			Help: "Number of FuzzyCronJob status update errors",
		}, []string{"cron_job_name", "namespace"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(CronCreation, CronCreationError, CronUpdate, CronUpdateError, FuzzyStatusUpdate, FuzzyStatusUpdateError)
}
