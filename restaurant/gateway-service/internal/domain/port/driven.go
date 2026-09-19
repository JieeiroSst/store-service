package port

import "net/http"

// Each proxy fronts one downstream service the gateway forwards to,
// matching the drawio's three outgoing edges from gateway-service. Distinct
// interfaces (rather than one shared type) let fx wire three independently
// configured *httputil.ReverseProxy instances without ambiguity.
type ConsumerProxy interface {
	http.Handler
}

type AccountingProxy interface {
	http.Handler
}

type DeliveryProxy interface {
	http.Handler
}
