package metrics

import (
	"net/http"

	"github.com/JIeeiroSst/auth-service/internal/port"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewPrometheus),
	fx.Provide(func(p *Prometheus) port.Metrics { return p }),
)

type Prometheus struct {
	registry   *prometheus.Registry
	challenges *prometheus.CounterVec
}

func NewPrometheus() *Prometheus {
	p := &Prometheus{
		registry: prometheus.NewRegistry(),
		challenges: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "payment_authentication_events_total",
			Help: "Payment authentication events: initiated, authenticated, otp_mismatch, failed, declined, consumed, consume_rejected.",
		}, []string{"event"}),
	}
	p.registry.MustRegister(p.challenges,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return p
}

func (p *Prometheus) ObserveChallenge(event string) { p.challenges.WithLabelValues(event).Inc() }

func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
