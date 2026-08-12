package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/dto"
	adminmw "github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type ProjectHandler struct {
	projects port.ProjectService
	rbac     port.RBACService
}

func NewProjectHandler(projects port.ProjectService, rbac port.RBACService) *ProjectHandler {
	return &ProjectHandler{projects: projects, rbac: rbac}
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProjectRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	userID, _ := adminmw.UserID(r.Context())

	p, err := h.projects.Create(r.Context(), req.Name, req.Description, userID)
	if err != nil {
		writeError(w, err)
		return
	}

	// The creator automatically becomes the project's Owner.
	if ownerRoleID, ok := h.findOwnerRoleID(r.Context()); ok {
		_, _ = h.rbac.AddMember(r.Context(), p.ID, userID, ownerRoleID)
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *ProjectHandler) findOwnerRoleID(ctx context.Context) (uuid.UUID, bool) {
	roles, err := h.rbac.ListRoles(ctx)
	if err != nil {
		return uuid.Nil, false
	}
	for _, role := range roles {
		if role.Name == model.RoleOwner {
			return role.ID, true
		}
	}
	return uuid.Nil, false
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	p, err := h.projects.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	var req dto.UpdateProjectRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, err)
		return
	}
	userID, _ := adminmw.UserID(r.Context())
	p, err := h.projects.Update(r.Context(), id, req.Name, req.Description, userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, apperr.ErrValidation)
		return
	}
	userID, _ := adminmw.UserID(r.Context())
	if err := h.projects.Delete(r.Context(), id, userID); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
