package http

import (
	"context"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
)

type (
	rolesOf          func(ctx context.Context, token, realm, id string) ([]*gocloak.Role, error)
	clientRolesOf    func(ctx context.Context, token, realm, idOfClient, id string) ([]*gocloak.Role, error)
	changeRoles      func(ctx context.Context, token, realm, id string, roles []gocloak.Role) error
	changeClientRole func(ctx context.Context, token, realm, idOfClient, id string, roles []gocloak.Role) error
)

type roleMappingOps struct {
	all             func(ctx context.Context, token, realm, id string) (*gocloak.MappingsRepresentation, error)
	realmGet        rolesOf
	realmComposite  rolesOf
	realmAvailable  rolesOf
	realmAdd        changeRoles
	realmDelete     changeRoles
	clientGet       clientRolesOf
	clientComposite clientRolesOf
	clientAvailable clientRolesOf
	clientAdd       changeClientRole
	clientDelete    changeClientRole
}

func (u *Handler) roleMappingRoutes(router chi.Router, idParam string, ops roleMappingOps) {
	getRoles := func(fn rolesOf) http.HandlerFunc {
		return u.admin(func(r *http.Request, token, realm string) (any, error) {
			return fn(r.Context(), token, realm, param(r, idParam))
		})
	}
	getClientRoles := func(fn clientRolesOf) http.HandlerFunc {
		return u.admin(func(r *http.Request, token, realm string) (any, error) {
			return fn(r.Context(), token, realm, param(r, "clientID"), param(r, idParam))
		})
	}
	setRoles := func(fn changeRoles) http.HandlerFunc {
		return u.admin(func(r *http.Request, token, realm string) (any, error) {
			roles, err := body[[]gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, fn(r.Context(), token, realm, param(r, idParam), roles)
		})
	}
	setClientRoles := func(fn changeClientRole) http.HandlerFunc {
		return u.admin(func(r *http.Request, token, realm string) (any, error) {
			roles, err := body[[]gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, fn(r.Context(), token, realm, param(r, "clientID"), param(r, idParam), roles)
		})
	}

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return ops.all(r.Context(), token, realm, param(r, idParam))
	}))
	router.Get("/realm", getRoles(ops.realmGet))
	router.Post("/realm", setRoles(ops.realmAdd))
	router.Delete("/realm", setRoles(ops.realmDelete))
	router.Get("/realm/composite", getRoles(ops.realmComposite))
	router.Get("/realm/available", getRoles(ops.realmAvailable))
	router.Get("/clients/{clientID}", getClientRoles(ops.clientGet))
	router.Post("/clients/{clientID}", setClientRoles(ops.clientAdd))
	router.Delete("/clients/{clientID}", setClientRoles(ops.clientDelete))
	router.Get("/clients/{clientID}/composite", getClientRoles(ops.clientComposite))
	router.Get("/clients/{clientID}/available", getClientRoles(ops.clientAvailable))
}

func (u *Handler) realmRoleRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetRoleParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetRealmRoles(r.Context(), token, realm, params)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Role](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateRealmRole(r.Context(), token, realm, req))
	}))

	router.Route("/{roleName}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetRealmRole(r.Context(), token, realm, param(r, "roleName"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, k.UpdateRealmRole(r.Context(), token, realm, param(r, "roleName"), req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteRealmRole(r.Context(), token, realm, param(r, "roleName"))
		}))
		router.Get("/composites", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetCompositeRealmRoles(r.Context(), token, realm, param(r, "roleName"))
		}))
		router.Post("/composites", u.admin(func(r *http.Request, token, realm string) (any, error) {
			roles, err := body[[]gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, k.AddRealmRoleComposite(r.Context(), token, realm, param(r, "roleName"), roles)
		}))
		router.Delete("/composites", u.admin(func(r *http.Request, token, realm string) (any, error) {
			roles, err := body[[]gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, k.DeleteRealmRoleComposite(r.Context(), token, realm, param(r, "roleName"), roles)
		}))
		router.Get("/users", u.admin(func(r *http.Request, token, realm string) (any, error) {
			params, err := query[gocloak.GetUsersByRoleParams](r)
			if err != nil {
				return nil, err
			}
			return k.GetUsersByRoleName(r.Context(), token, realm, param(r, "roleName"), params)
		}))
		router.Get("/groups", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetGroupsByRole(r.Context(), token, realm, param(r, "roleName"))
		}))
	})
}

func (u *Handler) roleByIDRoutes(router chi.Router) {
	k := u.keycloak

	router.Route("/{roleID}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetRealmRoleByID(r.Context(), token, realm, param(r, "roleID"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, k.UpdateRealmRoleByID(r.Context(), token, realm, param(r, "roleID"), req)
		}))
		router.Get("/composites", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetCompositeRolesByRoleID(r.Context(), token, realm, param(r, "roleID"))
		}))
		router.Post("/composites", u.admin(func(r *http.Request, token, realm string) (any, error) {
			roles, err := body[[]gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, k.AddClientRoleComposite(r.Context(), token, realm, param(r, "roleID"), roles)
		}))
		router.Delete("/composites", u.admin(func(r *http.Request, token, realm string) (any, error) {
			roles, err := body[[]gocloak.Role](r)
			if err != nil {
				return nil, err
			}
			return nil, k.DeleteClientRoleComposite(r.Context(), token, realm, param(r, "roleID"), roles)
		}))
		router.Get("/composites/realm", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetCompositeRealmRolesByRoleID(r.Context(), token, realm, param(r, "roleID"))
		}))
		router.Get("/composites/clients/{clientID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetCompositeClientRolesByRoleID(r.Context(), token, realm, param(r, "clientID"), param(r, "roleID"))
		}))
	})
}
