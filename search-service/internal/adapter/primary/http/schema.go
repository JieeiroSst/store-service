package http

import (
	"context"
	"net/http"
	"time"
)

const reindexTimeout = 30 * time.Minute

func (h *Handler) SchemaStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.schema.Status(r.Context())
	respond(w, r, status, err)
}

func (h *Handler) ApplySchema(w http.ResponseWriter, r *http.Request) {
	status, err := h.schema.Apply(r.Context())
	respond(w, r, status, err)
}

func (h *Handler) Reindex(w http.ResponseWriter, r *http.Request) {
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(reindexTimeout + time.Minute))
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), reindexTimeout)
	defer cancel()

	values := r.URL.Query()
	out, err := h.schema.Reindex(ctx, values["index"], values.Get("unmanaged") == "true")
	respond(w, r, out, err)
}
