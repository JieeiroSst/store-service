package http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/serpapi-service/config"
	"github.com/JIeeiroSst/serpapi-service/internal/application"
	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

const (
	maxBodyBytes         = 1 << 20
	maxMultipartOverhead = 64 << 10
)

type Handler struct {
	search  *application.SearchService
	account *application.AccountService
	images  *application.ImageService
	tokens  [][]byte
}

func NewHandler(cfg *config.Config, search *application.SearchService, account *application.AccountService, images *application.ImageService) *Handler {
	h := &Handler{search: search, account: account, images: images}
	for _, t := range cfg.Server.AccessTokens {
		h.tokens = append(h.tokens, []byte(t))
	}
	return h
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": config.Version})
}

func (h *Handler) Engines(w http.ResponseWriter, r *http.Request) {
	engines := h.search.Engines(r.URL.Query().Get("group"))
	writeJSON(w, http.StatusOK, map[string]any{"engines": engines, "count": len(engines)})
}

func (h *Handler) Engine(w http.ResponseWriter, r *http.Request) {
	e, err := h.search.Engine(r.PathValue("engine"))
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (h *Handler) SearchQuery(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	req := domain.SearchRequest{
		Engine:  q.Get("engine"),
		NoCache: isTrue(q.Get("no_cache")),
		Async:   isTrue(q.Get("async")),
		Params:  domain.Params{},
	}
	if e := r.PathValue("engine"); e != "" {
		req.Engine = e
	}
	output := q.Get("output")
	if output == "" && strings.Contains(r.Header.Get("Accept"), "text/markdown") {
		output = string(domain.OutputMarkdown)
	}
	out, err := domain.ParseOutput(output)
	if err != nil {
		h.fail(w, err)
		return
	}
	req.Output = out
	for k, v := range q {
		if len(v) > 0 {
			req.Params[k] = v[0]
		}
	}
	h.runSearch(w, r, req)
}

func (h *Handler) SearchBody(w http.ResponseWriter, r *http.Request) {
	var body searchRequestDTO
	if err := decodeBody(r, &body); err != nil {
		h.fail(w, err)
		return
	}
	req, err := body.toDomain()
	if err != nil {
		h.fail(w, err)
		return
	}
	h.runSearch(w, r, req)
}

func (h *Handler) runSearch(w http.ResponseWriter, r *http.Request, req domain.SearchRequest) {
	res, err := h.search.Search(r.Context(), req)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeResult(w, res)
}

func (h *Handler) SearchBatch(w http.ResponseWriter, r *http.Request) {
	var body batchRequestDTO
	if err := decodeBody(r, &body); err != nil {
		h.fail(w, err)
		return
	}
	reqs := make([]domain.SearchRequest, len(body.Searches))
	for i, s := range body.Searches {
		req, err := s.toDomain()
		if err == nil && !req.Output.IsJSON() {
			err = domain.Invalid("searches[%d]: batch supports json and json_with_pixel_position output only", i)
		}
		if err != nil {
			h.fail(w, err)
			return
		}
		reqs[i] = req
	}
	items, err := h.search.Batch(r.Context(), reqs)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := batchResponseDTO{Results: make([]batchItemDTO, len(items))}
	for i, it := range items {
		if it.Err != nil {
			status, msg := errorStatus(it.Err)
			out.Results[i] = batchItemDTO{Status: status, Error: msg}
			continue
		}
		out.Results[i] = batchItemDTO{
			Status:   http.StatusOK,
			SearchID: it.Result.SearchID,
			Cached:   it.Result.Cached,
			Data:     json.RawMessage(it.Result.Body),
		}
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	out, err := domain.ParseOutput(r.URL.Query().Get("output"))
	if err != nil {
		h.fail(w, err)
		return
	}
	res, err := h.search.Archive(r.Context(), r.PathValue("id"), out)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeResult(w, res)
}

func (h *Handler) Account(w http.ResponseWriter, r *http.Request) {
	acc, err := h.account.Account(r.Context())
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, acc)
}

func (h *Handler) Locations(w http.ResponseWriter, r *http.Request) {
	q := domain.LocationQuery{Q: r.URL.Query().Get("q")}
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			h.fail(w, domain.Invalid("limit must be a number"))
			return
		}
		q.Limit = n
	}
	locs, err := h.account.Locations(r.Context(), q)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, locs)
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, domain.MaxImageBytes+maxMultipartOverhead)
	var (
		name string
		data []byte
		err  error
	)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		name, data, err = readImagePart(r)
	} else {
		data, err = io.ReadAll(r.Body)
	}
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			err = domain.Invalid("image is larger than %d bytes", domain.MaxImageBytes)
		} else if !domain.IsInvalid(err) {
			err = domain.Invalid("read image: %v", err)
		}
		h.fail(w, err)
		return
	}
	out, err := h.images.Upload(r.Context(), name, data)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

func readImagePart(r *http.Request) (string, []byte, error) {
	mr, err := r.MultipartReader()
	if err != nil {
		return "", nil, err
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			return "", nil, domain.Invalid("multipart field \"image\" is required")
		}
		if err != nil {
			return "", nil, err
		}
		if part.FormName() != "image" {
			continue
		}
		data, err := io.ReadAll(part)
		return filepath.Base(part.FileName()), data, err
	}
}

func (h *Handler) authorize(next http.Handler) http.Handler {
	if len(h.tokens) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.validToken(token(r)) {
			w.Header().Set("WWW-Authenticate", `Bearer realm="serpapi-service"`)
			h.fail(w, domain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) validToken(t string) bool {
	if t == "" {
		return false
	}
	ok := 0
	for _, want := range h.tokens {
		ok |= subtle.ConstantTimeCompare([]byte(t), want)
	}
	return ok == 1
}

func token(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return r.Header.Get("X-Api-Token")
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	status, msg := errorStatus(err)
	if status >= 500 {
		log.Printf("request failed: %v", err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func errorStatus(err error) (int, string) {
	switch {
	case domain.IsInvalid(err):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "missing or invalid access token"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrExpired):
		return http.StatusGone, err.Error()
	case errors.Is(err, domain.ErrRateLimited):
		return http.StatusTooManyRequests, err.Error()
	case domain.IsUpstream(err):
		return http.StatusBadGateway, err.Error()
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "serpapi did not answer in time"
	case errors.Is(err, context.Canceled):
		return http.StatusServiceUnavailable, "request canceled"
	}
	return http.StatusInternalServerError, "internal error"
}

func writeResult(w http.ResponseWriter, res domain.SearchResult) {
	ct := res.ContentType
	if ct == "" {
		ct = "application/json; charset=utf-8"
	}
	w.Header().Set("Content-Type", ct)
	if res.Cached {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}
	if res.SearchID != "" {
		w.Header().Set("X-Search-Id", res.SearchID)
	}
	if res.Status != "" {
		w.Header().Set("X-Search-Status", res.Status)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(res.Body)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeBody(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return domain.Invalid("invalid JSON body: %v", err)
	}
	return nil
}

func isTrue(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
