package metrics

import (
	"net/http"

	"github.com/JIeeiroSst/webrtc-service/internal/domain"
	"github.com/JIeeiroSst/webrtc-service/internal/port"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Prometheus struct {
	registry *prometheus.Registry
	relayed  *prometheus.CounterVec
	rejected *prometheus.CounterVec
}

func NewPrometheus(rooms port.RoomRegistry) *Prometheus {
	p := &Prometheus{
		registry: prometheus.NewRegistry(),
		relayed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "webrtc_messages_relayed_total",
			Help: "Client messages relayed, by type.",
		}, []string{"type"}),
		rejected: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "webrtc_joins_rejected_total",
			Help: "Room joins refused, by reason.",
		}, []string{"reason"}),
	}
	p.registry.MustRegister(
		p.relayed,
		p.rejected,
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "webrtc_rooms",
			Help: "Rooms with at least one connected peer.",
		}, func() float64 { r, _ := rooms.Stats(); return float64(r) }),
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "webrtc_peers",
			Help: "Connected signaling clients.",
		}, func() float64 { _, n := rooms.Stats(); return float64(n) }),
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return p
}

func (p *Prometheus) MessageRelayed(t domain.MessageType) {
	p.relayed.WithLabelValues(string(t)).Inc()
}

func (p *Prometheus) JoinRejected(reason string) {
	p.rejected.WithLabelValues(reason).Inc()
}

func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
