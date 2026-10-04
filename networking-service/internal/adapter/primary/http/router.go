package http

import (
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/web"
)

func NewRouter(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ui/", http.StatusFound)
	})
	mux.HandleFunc("GET /ui/", h.UI)

	mux.HandleFunc("GET /v1/agent/self", h.AgentSelf)
	mux.HandleFunc("GET /v1/agent/members", h.AgentMembers)
	mux.HandleFunc("GET /v1/agent/services", h.AgentServices)
	mux.HandleFunc("GET /v1/agent/service/{id}", h.AgentService)
	mux.HandleFunc("PUT /v1/agent/service/register", h.AgentRegisterService)
	mux.HandleFunc("PUT /v1/agent/service/deregister/{id}", h.AgentDeregisterService)
	mux.HandleFunc("PUT /v1/agent/service/maintenance/{id}", h.AgentServiceMaintenance)
	mux.HandleFunc("PUT /v1/agent/maintenance", h.AgentNodeMaintenance)
	mux.HandleFunc("GET /v1/agent/checks", h.AgentChecks)
	mux.HandleFunc("PUT /v1/agent/check/register", h.AgentRegisterCheck)
	mux.HandleFunc("PUT /v1/agent/check/deregister/{id}", h.AgentDeregisterCheck)
	mux.HandleFunc("PUT /v1/agent/check/pass/{id}", h.AgentCheckTTL(domain.HealthPassing))
	mux.HandleFunc("PUT /v1/agent/check/warn/{id}", h.AgentCheckTTL(domain.HealthWarning))
	mux.HandleFunc("PUT /v1/agent/check/fail/{id}", h.AgentCheckTTL(domain.HealthCritical))
	mux.HandleFunc("PUT /v1/agent/check/update/{id}", h.AgentCheckUpdate)
	mux.HandleFunc("GET /v1/agent/health/service/name/{name}", h.AgentHealthByName)

	mux.HandleFunc("PUT /v1/catalog/register", h.CatalogRegister)
	mux.HandleFunc("PUT /v1/catalog/deregister", h.CatalogDeregister)
	mux.HandleFunc("GET /v1/catalog/datacenters", h.CatalogDatacenters)
	mux.HandleFunc("GET /v1/catalog/nodes", h.CatalogNodes)
	mux.HandleFunc("GET /v1/catalog/node/{node}", h.CatalogNode)
	mux.HandleFunc("GET /v1/catalog/services", h.CatalogServices)
	mux.HandleFunc("GET /v1/catalog/service/{service}", h.CatalogService)

	mux.HandleFunc("GET /v1/health/service/{service}", h.HealthService)
	mux.HandleFunc("GET /v1/health/checks/{service}", h.HealthServiceChecks)
	mux.HandleFunc("GET /v1/health/node/{node}", h.HealthNode)
	mux.HandleFunc("GET /v1/health/state/{state}", h.HealthState)

	mux.HandleFunc("GET /v1/kv/{key...}", h.KVGet)
	mux.HandleFunc("PUT /v1/kv/{key...}", h.KVPut)
	mux.HandleFunc("DELETE /v1/kv/{key...}", h.KVDelete)

	mux.HandleFunc("PUT /v1/session/create", h.SessionCreate)
	mux.HandleFunc("PUT /v1/session/destroy/{id}", h.SessionDestroy)
	mux.HandleFunc("PUT /v1/session/renew/{id}", h.SessionRenew)
	mux.HandleFunc("GET /v1/session/info/{id}", h.SessionInfo)
	mux.HandleFunc("GET /v1/session/list", h.SessionList)
	mux.HandleFunc("GET /v1/session/node/{node}", h.SessionList)

	mux.HandleFunc("GET /v1/connect/intentions", h.IntentionList)
	mux.HandleFunc("POST /v1/connect/intentions", h.IntentionCreate)
	mux.HandleFunc("GET /v1/connect/intentions/exact", h.IntentionGetExact)
	mux.HandleFunc("PUT /v1/connect/intentions/exact", h.IntentionPutExact)
	mux.HandleFunc("DELETE /v1/connect/intentions/exact", h.IntentionDeleteExact)
	mux.HandleFunc("GET /v1/connect/intentions/check", h.IntentionCheck)
	mux.HandleFunc("GET /v1/connect/intentions/match", h.IntentionMatch)
	mux.HandleFunc("GET /v1/connect/intentions/{id}", h.IntentionGet)
	mux.HandleFunc("PUT /v1/connect/intentions/{id}", h.IntentionUpdate)
	mux.HandleFunc("DELETE /v1/connect/intentions/{id}", h.IntentionDelete)

	mux.HandleFunc("GET /v1/status/leader", h.StatusLeader)
	mux.HandleFunc("GET /v1/status/peers", h.StatusPeers)
	mux.HandleFunc("GET /v1/snapshot", h.SnapshotSave)
	mux.HandleFunc("PUT /v1/snapshot", h.SnapshotRestore)
	mux.HandleFunc("GET /v1/internal/ui/services", h.UIServices)

	return mux
}

func (h *Handler) UI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(web.Index)
}
