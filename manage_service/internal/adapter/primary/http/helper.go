package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/JIeeiroSst/manage-service/internal/domain/port"
	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
)

var errMissingBearer = errors.New("missing bearer token")

type badRequestError struct{ err error }

func (e badRequestError) Error() string { return e.err.Error() }

type createdResult struct{ value any }

type handlerFunc func(r *http.Request) (any, error)

type adminFunc func(r *http.Request, token, realm string) (any, error)

func (u *Handler) handle(fn handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := fn(r)
		if err != nil {
			writeError(w, err)
			return
		}
		switch v := res.(type) {
		case nil:
			w.WriteHeader(http.StatusNoContent)
		case createdResult:
			writeJSON(w, http.StatusCreated, v.value)
		default:
			writeJSON(w, http.StatusOK, v)
		}
	}
}

func (u *Handler) admin(fn adminFunc) http.HandlerFunc {
	return u.handle(func(r *http.Request) (any, error) {
		token, err := u.adminToken(r)
		if err != nil {
			return nil, err
		}
		return fn(r, token, chi.URLParam(r, "realm"))
	})
}

func (u *Handler) user(fn adminFunc) http.HandlerFunc {
	return u.handle(func(r *http.Request) (any, error) {
		token := bearerToken(r)
		if token == "" {
			return nil, errMissingBearer
		}
		return fn(r, token, chi.URLParam(r, "realm"))
	})
}

func (u *Handler) adminToken(r *http.Request) (string, error) {
	if token := bearerToken(r); token != "" {
		return token, nil
	}
	return u.keycloak.AdminToken(r.Context())
}

func bearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if len(auth) > 7 && strings.EqualFold(auth[:7], "bearer ") {
		return strings.TrimSpace(auth[7:])
	}
	return ""
}

func param(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

func body[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, badRequestError{err}
	}
	return v, nil
}

func query[T any](r *http.Request, arrayKeys ...string) (T, error) {
	var v T
	values := map[string]any{}
	for key, vals := range r.URL.Query() {
		if len(vals) > 1 || contains(arrayKeys, key) {
			values[key] = vals
		} else {
			values[key] = vals[0]
		}
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return v, badRequestError{err}
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, badRequestError{err}
	}
	return v, nil
}

func withoutKey(values url.Values, key string) string {
	values.Del(key)
	return values.Encode()
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func created(id string, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	return createdResult{CreatedResponse{ID: id}}, nil
}

func createdValue[T any](v T, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	return createdResult{v}, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	var apiErr *gocloak.APIError
	var badReq badRequestError
	switch {
	case errors.As(err, &apiErr) && apiErr.Code > 0:
		status = apiErr.Code
	case errors.As(err, &badReq):
		status = http.StatusBadRequest
	case errors.Is(err, errMissingBearer), errors.Is(err, port.ErrAdminNotConfigured):
		status = http.StatusUnauthorized
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
