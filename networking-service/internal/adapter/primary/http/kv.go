package http

import (
	"io"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func (h *Handler) KVGet(w http.ResponseWriter, r *http.Request) {
	q, ok := h.query(w, r)
	if !ok {
		return
	}
	key := r.PathValue("key")
	switch {
	case r.URL.Query().Has("keys"):
		keys, m, err := h.kv.Keys(r.Context(), q, key, r.URL.Query().Get("separator"))
		if err != nil {
			h.fail(w, err)
			return
		}
		if len(keys) == 0 {
			h.setMeta(w, m)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.ok(w, r, &m, keys)

	case flag(r, "recurse"):
		pairs, m, err := h.kv.List(r.Context(), q, key)
		if err != nil {
			h.fail(w, err)
			return
		}
		if len(pairs) == 0 {
			h.setMeta(w, m)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		out := make([]kvOut, 0, len(pairs))
		for _, p := range pairs {
			out = append(out, toKVOut(p))
		}
		h.ok(w, r, &m, out)

	default:
		if key == "" {
			h.fail(w, domain.Invalid("missing key name"))
			return
		}
		pair, m, err := h.kv.Get(r.Context(), q, key)
		if err != nil {
			h.fail(w, err)
			return
		}
		if pair == nil {
			h.setMeta(w, m)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if flag(r, "raw") {
			h.setMeta(w, m)
			ct := "application/octet-stream"
			if utf8.Valid(pair.Value) {
				ct = "text/plain; charset=utf-8"
			}
			w.Header().Set("Content-Type", ct)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			_, _ = w.Write(pair.Value)
			return
		}
		h.ok(w, r, &m, []kvOut{toKVOut(*pair)})
	}
}

func (h *Handler) KVPut(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	q := r.URL.Query()
	pair := domain.KVPair{Key: r.PathValue("key")}
	var err error
	if v := q.Get("flags"); v != "" {
		if pair.Flags, err = strconv.ParseUint(v, 10, 64); err != nil {
			h.fail(w, domain.Invalid("invalid flags"))
			return
		}
	}
	cas, err := casParam(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	value, err := io.ReadAll(io.LimitReader(r.Body, int64(h.kv.MaxValueBytes())+1))
	if err != nil {
		h.fail(w, domain.Invalid("read body: %v", err))
		return
	}
	pair.Value = value
	done, err := h.kv.Put(pair, cas, q.Get("acquire"), q.Get("release"))
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, done)
}

func (h *Handler) KVDelete(w http.ResponseWriter, r *http.Request) {
	if !h.write(w, r) {
		return
	}
	cas, err := casParam(r)
	if err != nil {
		h.fail(w, err)
		return
	}
	done, err := h.kv.Delete(r.PathValue("key"), flag(r, "recurse"), cas)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.ok(w, r, nil, done)
}

func casParam(r *http.Request) (*uint64, error) {
	v := r.URL.Query().Get("cas")
	if v == "" {
		return nil, nil
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return nil, domain.Invalid("invalid cas index")
	}
	return &n, nil
}
