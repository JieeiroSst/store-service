package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func (h *Handler) StatusLeader(w http.ResponseWriter, r *http.Request) {
	h.ok(w, r, nil, leaderAddr(h.self.Address))
}

func (h *Handler) StatusPeers(w http.ResponseWriter, r *http.Request) {
	h.ok(w, r, nil, []string{leaderAddr(h.self.Address)})
}

func (h *Handler) SnapshotSave(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	snap, err := h.snapshots.Export()
	if err != nil {
		h.fail(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="networking-snapshot.json"`)
	h.ok(w, r, &queryMeta{LastIndex: snap.Index}, snap)
}

func (h *Handler) SnapshotRestore(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	var snap domain.Snapshot
	if err := json.NewDecoder(io.LimitReader(r.Body, 512<<20)).Decode(&snap); err != nil {
		h.fail(w, domain.Invalid("decode snapshot: %v", err))
		return
	}
	if err := h.snapshots.Import(snap); err != nil {
		h.fail(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) UIServices(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	list, m, err := h.catalog.ServiceSummaries(r.Context(), q)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, &m, list)
}
