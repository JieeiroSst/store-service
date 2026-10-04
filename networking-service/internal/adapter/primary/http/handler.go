package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/application"
	"github.com/JIeeiroSst/networking-service/internal/domain"
)

const maxBodyBytes = 1 << 20

type (
	queryOpts = application.QueryOptions
	queryMeta = application.QueryMeta
)

type Handler struct {
	catalog    *application.CatalogService
	kv         *application.KVService
	sessions   *application.SessionService
	intentions *application.IntentionService
	snapshots  *application.Snapshotter
	acl        *application.Authorizer
	self       application.AgentInfo
}

func NewHandler(
	catalog *application.CatalogService,
	kv *application.KVService,
	sessions *application.SessionService,
	intentions *application.IntentionService,
	snapshots *application.Snapshotter,
	acl *application.Authorizer,
) *Handler {
	return &Handler{
		catalog: catalog, kv: kv, sessions: sessions, intentions: intentions,
		snapshots: snapshots, acl: acl, self: catalog.Self(),
	}
}

func token(r *http.Request) string {
	if t := r.Header.Get("X-Consul-Token"); t != "" {
		return t
	}
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return r.URL.Query().Get("token")
}

func (h *Handler) guard(w http.ResponseWriter, r *http.Request, write bool) bool {
	if dc := r.URL.Query().Get("dc"); dc != "" && dc != h.self.Datacenter {
		http.Error(w, "No path to datacenter", http.StatusInternalServerError)
		return false
	}
	if err := h.acl.Authorize(token(r), write); err != nil {
		h.fail(w, err)
		return false
	}
	return true
}

func (h *Handler) read(w http.ResponseWriter, r *http.Request) bool  { return h.guard(w, r, false) }
func (h *Handler) write(w http.ResponseWriter, r *http.Request) bool { return h.guard(w, r, true) }

func queryOptions(r *http.Request) (application.QueryOptions, error) {
	q := r.URL.Query()
	var opts application.QueryOptions
	if q.Get("filter") != "" {
		return opts, domain.Invalid("filter expressions are not supported")
	}
	if v := q.Get("index"); v != "" {
		idx, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return opts, domain.Invalid("Invalid index")
		}
		opts.MinIndex = idx
	}
	if v := q.Get("wait"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return opts, domain.Invalid("Invalid wait time")
		}
		opts.Wait = d
	}
	for _, kv := range q["node-meta"] {
		k, v, _ := strings.Cut(kv, ":")
		if opts.NodeMeta == nil {
			opts.NodeMeta = map[string]string{}
		}
		opts.NodeMeta[k] = v
	}
	return opts, nil
}

func flag(r *http.Request, name string) bool {
	q := r.URL.Query()
	if !q.Has(name) {
		return false
	}
	return q.Get(name) != "false"
}

func decode(r *http.Request, v any) error {
	return decodeOptional(r, v, false)
}

func decodeOptional(r *http.Request, v any, optional bool) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
	if err != nil {
		return domain.Invalid("read body: %v", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		if optional {
			return nil
		}
		return domain.Invalid("request body is required")
	}
	if err := json.Unmarshal(body, v); err != nil {
		if domain.IsInvalid(err) {
			return err
		}
		return domain.Invalid("Request decode failed: %v", err)
	}
	return nil
}

func (h *Handler) setMeta(w http.ResponseWriter, m application.QueryMeta) {
	hd := w.Header()
	hd.Set("X-Consul-Index", strconv.FormatUint(m.LastIndex, 10))
	hd.Set("X-Consul-KnownLeader", "true")
	hd.Set("X-Consul-LastContact", "0")
	hd.Set("X-Consul-Default-ACL-Policy", h.acl.DefaultPolicy())
}

func (h *Handler) reply(w http.ResponseWriter, r *http.Request, m *application.QueryMeta, status int, v any) {
	if m != nil {
		h.setMeta(w, *m)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	if flag(r, "pretty") {
		enc.SetIndent("", "    ")
	}
	if err := enc.Encode(v); err != nil {
		log.Printf("http: encode response: %v", err)
	}
}

func (h *Handler) ok(w http.ResponseWriter, r *http.Request, m *application.QueryMeta, v any) {
	h.reply(w, r, m, http.StatusOK, v)
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	var tooLarge application.ErrValueTooLarge
	switch {
	case domain.IsInvalid(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrPermissionDenied):
		http.Error(w, "Permission denied", http.StatusForbidden)
	case errors.Is(err, domain.ErrInvalidSession):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.As(err, &tooLarge):
		http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
	default:
		log.Printf("http: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"status":"ok"}`)
}
