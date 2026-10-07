package metrics

import (
	"net/http"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"github.com/JIeeiroSst/customer-info-service/internal/port"
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
	onboarding *prometheus.CounterVec
	kyc        *prometheus.CounterVec
}

func NewPrometheus() *Prometheus {
	p := &Prometheus{
		registry: prometheus.NewRegistry(),
		onboarding: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "customer_onboarding_total",
			Help: "Customer onboarding requests, by result (created, existing, failed).",
		}, []string{"result"}),
		kyc: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "customer_kyc_submissions_total",
			Help: "KYC document submissions, by resulting status.",
		}, []string{"status"}),
	}
	p.registry.MustRegister(p.onboarding, p.kyc,
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return p
}

func (p *Prometheus) ObserveOnboarding(result string) { p.onboarding.WithLabelValues(result).Inc() }

func (p *Prometheus) ObserveKYC(status domain.KYCStatus) { p.kyc.WithLabelValues(string(status)).Inc() }

func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
