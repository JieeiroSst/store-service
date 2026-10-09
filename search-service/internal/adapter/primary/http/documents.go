package http

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/search-service/internal/domain"
)

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q, err := parseDocumentQuery(r.URL.Query())
	if err != nil {
		writeError(w, r, err)
		return
	}
	page, err := h.documents.Search(r.Context(), q)
	respond(w, r, page, err)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q, err := parseDocumentQuery(r.URL.Query())
	if err != nil {
		writeError(w, r, err)
		return
	}
	page, err := h.documents.List(r.Context(), q)
	respond(w, r, page, err)
}

func (h *Handler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	size, err := parseSize(values)
	if err != nil {
		writeError(w, r, err)
		return
	}
	docs, err := h.documents.Autocomplete(r.Context(), domain.AutocompleteQuery{
		Keyword: values.Get("q"),
		Indices: values["index"],
		Size:    size,
	})
	respond(w, r, docs, err)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	doc, err := h.documents.Get(r.Context(), documentRef(r))
	respond(w, r, doc, err)
}

func (h *Handler) Similar(w http.ResponseWriter, r *http.Request) {
	size, err := parseSize(r.URL.Query())
	if err != nil {
		writeError(w, r, err)
		return
	}
	docs, err := h.documents.Similar(r.Context(), documentRef(r), size)
	respond(w, r, docs, err)
}

func (h *Handler) Indices(w http.ResponseWriter, r *http.Request) {
	indices, err := h.documents.Indices(r.Context())
	respond(w, r, indices, err)
}

func (h *Handler) Fields(w http.ResponseWriter, r *http.Request) {
	fields, err := h.documents.Fields(r.Context(), r.PathValue("index"))
	respond(w, r, fields, err)
}

func respond(w http.ResponseWriter, r *http.Request, data any, err error) {
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func documentRef(r *http.Request) domain.DocumentRef {
	return domain.DocumentRef{Index: r.PathValue("index"), ID: r.PathValue("id")}
}

func parseDocumentQuery(values url.Values) (domain.DocumentQuery, error) {
	size, err := parseSize(values)
	if err != nil {
		return domain.DocumentQuery{}, err
	}

	filters := map[string][]string{}
	ranges := map[string]domain.Range{}
	for key, vals := range values {
		op, field, ok := bracketKey(key)
		if !ok || len(vals) == 0 {
			continue
		}
		value := vals[0]
		if op == "filter" {
			filters[field] = append(filters[field], vals...)
			continue
		}
		rng := ranges[field]
		switch op {
		case "gt":
			rng.Gt = value
		case "gte":
			rng.Gte = value
		case "lt":
			rng.Lt = value
		case "lte":
			rng.Lte = value
		default:
			continue
		}
		ranges[field] = rng
	}

	return domain.DocumentQuery{
		Keyword: values.Get("q"),
		Indices: values["index"],
		Filters: filters,
		Ranges:  ranges,
		Sort:    domain.ParseSort(values.Get("sort")),
		Facets:  values["facets"],
		Size:    size,
		Cursor:  values.Get("cursor"),
	}, nil
}

func bracketKey(key string) (string, string, bool) {
	open := strings.IndexByte(key, '[')
	if open <= 0 || !strings.HasSuffix(key, "]") {
		return "", "", false
	}
	field := key[open+1 : len(key)-1]
	if field == "" {
		return "", "", false
	}
	return key[:open], field, true
}

func parseSize(values url.Values) (int, error) {
	raw := values.Get("size")
	if raw == "" {
		return 0, nil
	}
	size, err := strconv.Atoi(raw)
	if err != nil || size <= 0 {
		return 0, domain.Invalid("size must be a positive integer")
	}
	return size, nil
}
