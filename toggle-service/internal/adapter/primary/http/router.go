package httpadapter

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/handler"
	appmw "github.com/JIeeiroSst/toggle-service/internal/adapter/primary/http/middleware"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type RouterParams struct {
	fx.In

	Logger *zap.Logger

	AuthService  port.AuthService
	RBACService  port.RBACService
	TokenService port.TokenService

	Projects     *handler.ProjectHandler
	Environments *handler.EnvironmentHandler
	Flags        *handler.FeatureFlagHandler
	Strategies   *handler.StrategyHandler
	Auth         *handler.AuthHandler
	RBAC         *handler.RBACHandler
	Tokens       *handler.TokenHandler
	Audit        *handler.AuditHandler
	Client       *handler.ClientHandler
}

func NewRouter(p RouterParams) chi.Router {
	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.Recoverer)
	r.Use(zapRequestLogger(p.Logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	adminAuth := appmw.RequireAdminAuth(p.AuthService)

	r.Route("/api/admin", func(admin chi.Router) {
		admin.Post("/auth/register", p.Auth.Register)
		admin.Post("/auth/login", p.Auth.Login)

		admin.Group(func(protected chi.Router) {
			protected.Use(adminAuth)

			protected.Get("/roles", p.RBAC.ListRoles)

			protected.Route("/environments", func(env chi.Router) {
				env.Use(appmw.RequireInstanceAdmin)
				env.Get("/", p.Environments.List)
				env.Post("/", p.Environments.Create)
				env.Put("/{environmentId}", p.Environments.Update)
				env.Delete("/{environmentId}", p.Environments.Delete)
			})

			protected.Route("/tokens", func(tok chi.Router) {
				tok.Use(appmw.RequireInstanceAdmin)
				tok.Get("/", p.Tokens.List)
				tok.Post("/", p.Tokens.Create)
				tok.Delete("/{tokenId}", p.Tokens.Revoke)
			})

			protected.Route("/members/{membershipId}", func(m chi.Router) {
				m.Use(appmw.RequireInstanceAdmin)
				m.Put("/", p.RBAC.UpdateMember)
				m.Delete("/", p.RBAC.RemoveMember)
			})

			protected.Route("/projects", func(proj chi.Router) {
				proj.Get("/", p.Projects.List)
				proj.Post("/", p.Projects.Create)

				proj.Route("/{projectId}", func(pr chi.Router) {
					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionView)).Get("/", p.Projects.Get)
					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionManageProject)).Put("/", p.Projects.Update)
					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionManageProject)).Delete("/", p.Projects.Delete)

					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionManageMembers)).Get("/members", p.RBAC.ListMembers)
					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionManageMembers)).Post("/members", p.RBAC.AddMember)

					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionView)).Get("/audit", p.Audit.List)

					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionView)).Get("/flags", p.Flags.List)
					pr.With(appmw.RequirePermission(p.RBACService, model.PermissionCreateFeature)).Post("/flags", p.Flags.Create)

					pr.Route("/flags/{key}", func(f chi.Router) {
						f.With(appmw.RequirePermission(p.RBACService, model.PermissionView)).Get("/", p.Flags.Get)
						f.With(appmw.RequirePermission(p.RBACService, model.PermissionUpdateFeature)).Put("/", p.Flags.Update)
						f.With(appmw.RequirePermission(p.RBACService, model.PermissionDeleteFeature)).Delete("/", p.Flags.Archive)

						f.Route("/environments/{envName}", func(fe chi.Router) {
							fe.With(appmw.RequirePermission(p.RBACService, model.PermissionToggleFeature)).Patch("/on", p.Flags.ToggleOn)
							fe.With(appmw.RequirePermission(p.RBACService, model.PermissionToggleFeature)).Patch("/off", p.Flags.ToggleOff)

							fe.With(appmw.RequirePermission(p.RBACService, model.PermissionView)).Get("/strategies", p.Strategies.List)
							fe.With(appmw.RequirePermission(p.RBACService, model.PermissionCreateStrategy)).Post("/strategies", p.Strategies.Create)
							fe.With(appmw.RequirePermission(p.RBACService, model.PermissionCreateStrategy)).Put("/strategies/{strategyId}", p.Strategies.Update)
							fe.With(appmw.RequirePermission(p.RBACService, model.PermissionCreateStrategy)).Delete("/strategies/{strategyId}", p.Strategies.Delete)
						})
					})
				})
			})
		})
	})

	clientAuth := appmw.RequireClientToken(p.TokenService)
	r.Route("/api/client", func(c chi.Router) {
		c.Use(clientAuth)
		c.Get("/features", p.Client.GetFeatures)
		c.Post("/metrics", p.Client.Metrics)
		c.Post("/evaluate", p.Client.Evaluate)
	})

	return r
}

func zapRequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("http request", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			next.ServeHTTP(w, r)
		})
	}
}
