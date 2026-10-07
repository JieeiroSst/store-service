package metrics

import (
	"net/http"

	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	registry  *prometheus.Registry
	issued    *prometheus.CounterVec
	lifecycle *prometheus.CounterVec
	auths     *prometheus.CounterVec
}

func NewPrometheus() *Prometheus {
	p := &Prometheus{
		registry: prometheus.NewRegistry(),
		issued: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "card_issued_total",
			Help: "Cards issued, by program and card type.",
		}, []string{"program", "type"}),
		lifecycle: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "card_lifecycle_events_total",
			Help: "Account, card and authorization lifecycle changes, by event.",
		}, []string{"event"}),
		auths: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "card_authorizations_total",
			Help: "Authorization decisions, by channel and ISO 8583 response code.",
		}, []string{"channel", "code"}),
	}
	p.registry.MustRegister(
		p.issued, p.lifecycle, p.auths,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return p
}

func (p *Prometheus) ObserveIssued(program string, cardType domain.CardType) {
	p.issued.WithLabelValues(program, string(cardType)).Inc()
}

func (p *Prometheus) ObserveLifecycle(event string) {
	p.lifecycle.WithLabelValues(event).Inc()
}

func (p *Prometheus) ObserveAuthorization(channel string, code domain.ResponseCode) {
	p.auths.WithLabelValues(channel, string(code)).Inc()
}

func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
