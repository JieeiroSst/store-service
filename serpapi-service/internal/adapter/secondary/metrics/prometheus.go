package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	registry *prometheus.Registry
	calls    *prometheus.CounterVec
	latency  *prometheus.HistogramVec
	cache    *prometheus.CounterVec
}

func NewPrometheus() *Prometheus {
	p := &Prometheus{
		registry: prometheus.NewRegistry(),
		calls: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "serpapi_upstream_requests_total",
			Help: "Calls to SerpApi, by operation, engine and outcome.",
		}, []string{"op", "engine", "outcome"}),
		latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "serpapi_upstream_request_duration_seconds",
			Help:    "Latency of calls to SerpApi, by operation.",
			Buckets: []float64{.25, .5, 1, 2, 4, 8, 16, 32, 64},
		}, []string{"op"}),
		cache: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "serpapi_cache_lookups_total",
			Help: "Cache lookups, by operation and result.",
		}, []string{"op", "hit"}),
	}
	p.registry.MustRegister(
		p.calls, p.latency, p.cache,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return p
}

func (p *Prometheus) ObserveUpstream(op, engine, outcome string, d time.Duration) {
	p.calls.WithLabelValues(op, engine, outcome).Inc()
	p.latency.WithLabelValues(op).Observe(d.Seconds())
}

func (p *Prometheus) ObserveCache(op string, hit bool) {
	p.cache.WithLabelValues(op, strconv.FormatBool(hit)).Inc()
}

func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
