package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/partner-service/internal/adapters/handler"
	"github.com/JIeeiroSst/partner-service/internal/core/domain"
	"github.com/JIeeiroSst/partner-service/internal/core/ports"
	"github.com/JIeeiroSst/partner-service/internal/core/services"
	"github.com/gin-gonic/gin"
)

type partnerRepo struct{ ports.PartnerRepository }

func (partnerRepo) ReadPartner(id string) (*domain.Partner, error) {
	return &domain.Partner{ID: id}, nil
}

type partnershipRepo struct{ ports.PartnershipRepository }

func (partnershipRepo) ReadPartnership(id string) (*domain.Partnership, error) {
	return &domain.Partnership{ID: id}, nil
}

type partnershipsPartnerRepo struct {
	ports.PartnershipsPartnerRepository
}

func (partnershipsPartnerRepo) ReadPartnershipsPartner(id string) (*domain.PartnershipsPartner, error) {
	return &domain.PartnershipsPartner{ID: id}, nil
}

type projectRepo struct{ ports.ProjectRepository }

func (projectRepo) ReadProject(id string) (*domain.Project, error) {
	return &domain.Project{ID: id}, nil
}

func (projectRepo) ReadProjects(p domain.Pagination) (*domain.Pagination, error) {
	p.Rows = []domain.Project{{ID: "p1"}, {ID: "p2"}}
	p.TotalRows = 2
	return &p, nil
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(
		handler.NewPartnerHandler(*services.NewPartnerService(partnerRepo{})),
		handler.NewPartnershipHandler(*services.NewPartnershipService(partnershipRepo{})),
		handler.NewPartnershipsPartnerHandler(*services.NewPartnershipsPartnerService(partnershipsPartnerRepo{})),
		handler.NewProjectHandler(*services.NewProjectService(projectRepo{})),
	)
}

func get(t *testing.T, r http.Handler, path string, out any) {
	t.Helper()
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d %s", path, rec.Code, rec.Body)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
}

func TestGetByIDReadsThePathParameter(t *testing.T) {
	r := newTestRouter()
	for _, path := range []string{
		"/v1/partner/abc",
		"/v1/partnership/abc",
		"/v1/partnerships-partner/abc",
		"/v1/project/abc",
	} {
		var got struct {
			ID string `json:"id"`
		}
		get(t, r, path, &got)
		if got.ID != "abc" {
			t.Errorf("GET %s returned id %q, want abc", path, got.ID)
		}
	}
}

func TestProjectListIsAtTheCollectionRoot(t *testing.T) {
	var page struct {
		TotalRows int64            `json:"total_rows"`
		Rows      []map[string]any `json:"rows"`
	}
	get(t, newTestRouter(), "/v1/project/?limit=10", &page)
	if page.TotalRows != 2 || len(page.Rows) != 2 {
		t.Fatalf("GET /v1/project/ = %+v, want the paged list", page)
	}
}
