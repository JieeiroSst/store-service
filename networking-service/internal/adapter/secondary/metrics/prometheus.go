package metrics

import (
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	registry    *prometheus.Registry
	checks      *prometheus.CounterVec
	invalidated *prometheus.CounterVec
	woken       prometheus.Counter
}

func NewPrometheus(store port.StateStore) *Prometheus {
	p := &Prometheus{
		registry: prometheus.NewRegistry(),
		checks: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "networking_check_runs_total",
			Help: "Health check executions, by check type and resulting status.",
		}, []string{"type", "status"}),
		invalidated: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "networking_sessions_invalidated_total",
			Help: "Sessions invalidated, by reason.",
		}, []string{"reason"}),
		woken: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "networking_blocking_query_wakeups_total",
			Help: "Times a blocking query was woken by a write.",
		}),
	}
	stats := prometheus.NewDesc("networking_catalog_objects", "Objects in the state store, by kind.", []string{"kind"}, nil)
	checkStates := prometheus.NewDesc("networking_health_checks", "Health checks, by status.", []string{"status"}, nil)
	index := prometheus.NewDesc("networking_state_index", "Index of the latest write to the state store.", nil, nil)
	p.registry.MustRegister(
		p.checks,
		p.invalidated,
		p.woken,
		collectorFunc{
			descs: []*prometheus.Desc{stats, checkStates, index},
			collect: func(ch chan<- prometheus.Metric) {
				st, err := store.Stats()
				if err != nil {
					return
				}
				for kind, n := range map[string]int{
					"nodes": st.Nodes, "services": st.Services, "instances": st.Instances,
					"keys": st.Keys, "sessions": st.Sessions, "intentions": st.Intentions,
				} {
					ch <- prometheus.MustNewConstMetric(stats, prometheus.GaugeValue, float64(n), kind)
				}
				for _, s := range []domain.HealthStatus{domain.HealthPassing, domain.HealthWarning, domain.HealthCritical} {
					ch <- prometheus.MustNewConstMetric(checkStates, prometheus.GaugeValue, float64(st.Checks[s]), string(s))
				}
				if last, err := store.LastIndex(); err == nil {
					ch <- prometheus.MustNewConstMetric(index, prometheus.GaugeValue, float64(last))
				}
			},
		},
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return p
}

func (p *Prometheus) CheckRun(t domain.CheckType, s domain.HealthStatus) {
	p.checks.WithLabelValues(string(t), string(s)).Inc()
}

func (p *Prometheus) SessionInvalidated(reason string) {
	p.invalidated.WithLabelValues(reason).Inc()
}

func (p *Prometheus) BlockingQueryWoken() { p.woken.Inc() }

func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

type collectorFunc struct {
	descs   []*prometheus.Desc
	collect func(chan<- prometheus.Metric)
}

func (c collectorFunc) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descs {
		ch <- d
	}
}

func (c collectorFunc) Collect(ch chan<- prometheus.Metric) { c.collect(ch) }
