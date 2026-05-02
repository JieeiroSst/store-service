package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobsScheduled = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "notifyhub_jobs_scheduled_total",
		Help: "Number of jobs currently registered in the scheduler.",
	})

	JobExecutionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifyhub_job_executions_total",
		Help: "Total job execution attempts.",
	}, []string{"schedule_type", "status"}) // status: success|fail

	NotificationsSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifyhub_notifications_sent_total",
		Help: "Total notifications dispatched.",
	}, []string{"channel", "status"}) // status: sent|failed

	WorkerQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "notifyhub_worker_queue_depth",
		Help: "Current number of tasks waiting in the worker queue.",
	})

	FetchDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "notifyhub_fetch_duration_seconds",
		Help:    "Latency of external data source HTTP calls.",
		Buckets: prometheus.DefBuckets,
	}, []string{"source_name", "status"}) // status: ok|error

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "notifyhub_http_request_duration_seconds",
		Help:    "HTTP request latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})

	RetryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "notifyhub_retry_total",
		Help: "Total notification retry attempts.",
	}, []string{"channel"})
)
