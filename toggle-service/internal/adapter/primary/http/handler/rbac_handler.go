package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type RBACHandler struct {
	rbac port.RBACService
}

func NewRBACHandler(rbac port.RBACService) *RBACHandler {
	return &RBACHandler{rbac: rbac}
}

func (h *RBACHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.rbac.ListRoles(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func (h *RBACHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	members, err := h.rbac.ListMembers(r.Context(), projectID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func (h *RBACHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDParam(r)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.AddMemberRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	// req.UserID is a user_service user ID (opaque external identifier),
	// not a local UUID — passed through as-is.
	m, err := h.rbac.AddMember(r.Context(), projectID, req.UserID, roleID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *RBACHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	membershipID, err := uuid.Parse(chi.URLParam(r, "membershipId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.UpdateMemberRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	if err := h.rbac.UpdateMemberRole(r.Context(), membershipID, roleID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RBACHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	membershipID, err := uuid.Parse(chi.URLParam(r, "membershipId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	if err := h.rbac.RemoveMember(r.Context(), membershipID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
