package http

import (
	"net/http"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func (h *Handler) IntentionList(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	list, m, err := h.intentions.List(r.Context(), q)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, &m, intentionsOut(list))
}

func (h *Handler) IntentionCreate(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	in, err := decodeIntention(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	id, err := h.intentions.Create(in)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, map[string]string{"ID": id})
}

func (h *Handler) IntentionGet(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	in, err := h.intentions.Get(r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, toIntentionOut(*in))
}

func (h *Handler) IntentionUpdate(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	in, err := decodeIntention(r)
	if err == nil {
		err = h.intentions.Update(r.PathValue("id"), in)
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func (h *Handler) IntentionDelete(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	if err := h.intentions.Delete(r.PathValue("id")); err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func exactParams(r *http.Request) (string, string, error) {
	src, dst := r.URL.Query().Get("source"), r.URL.Query().Get("destination")
	if src == "" || dst == "" {
		return "", "", domain.Invalid("required query parameters 'source' and 'destination' not set")
	}
	return src, dst, nil
}

func (h *Handler) IntentionGetExact(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	src, dst, err := exactParams(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	in, err := h.intentions.GetExact(src, dst)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, toIntentionOut(*in))
}

func (h *Handler) IntentionPutExact(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	src, dst, err := exactParams(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	in, err := decodeIntention(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	in.SourceName, in.DestinationName = src, dst
	if err := h.intentions.UpsertExact(in); err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func (h *Handler) IntentionDeleteExact(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	src, dst, err := exactParams(r)
	if err == nil {
		err = h.intentions.DeleteExact(src, dst)
	}
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, true)
}

func (h *Handler) IntentionCheck(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	allowed, _, err := h.intentions.Check(r.URL.Query().Get("source"), r.URL.Query().Get("destination"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, map[string]bool{"Allowed": allowed})
}

func (h *Handler) IntentionMatch(w http.ResponseWriter, r *http.Request) {
	if !h.read(w, r) {
		return
	}
	by := r.URL.Query().Get("by")
	out := map[string][]intentionOut{}
	for _, name := range r.URL.Query()["name"] {
		list, err := h.intentions.Match(by, name)
		if err != nil {
			h.fail(w, err)
			return
		}
		out[name] = intentionsOut(list)
	}
	h.ok(w, r, nil, out)
}

func decodeIntention(r *http.Request) (domain.Intention, error) {
	var in intentionIn
	if err := decode(r, &in); err != nil {
		return domain.Intention{}, err
	}
	return in.toDomain()
}
