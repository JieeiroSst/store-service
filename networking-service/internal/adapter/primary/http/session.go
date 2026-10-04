package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func (h *Handler) SessionCreate(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var in sessionIn
	if err := decodeOptional(r, &in, true); err != nil {
		h.fail(w, err)
		return
	}
	sess, err := in.toDomain()
	if err != nil {
		h.fail(w, err)
		return
	}
	id, err := h.sessions.Create(sess)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, map[string]string{"ID": id})
}

func (h *Handler) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	if err := h.sessions.Destroy(r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func (h *Handler) SessionRenew(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	id := r.PathValue("id")
	sess, err := h.sessions.Renew(id)
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(w, fmt.Sprintf("Session id '%s' not found", id), http.StatusNotFound)
		return
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, []sessionOut{toSessionOut(*sess)})
}

func (h *Handler) SessionInfo(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	sess, m, err := h.sessions.Info(r.Context(), q, r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := []sessionOut{}
	if sess != nil {
		out = append(out, toSessionOut(*sess))
	}
	h.ok(w, r, &m, out)
}

func (h *Handler) SessionList(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	list, m, err := h.sessions.List(r.Context(), q, r.PathValue("node"))
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]sessionOut, 0, len(list))
	for _, s := range list {
		out = append(out, toSessionOut(s))
	}
	h.ok(w, r, &m, out)
}
