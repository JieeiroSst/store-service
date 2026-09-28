package http

import (
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
)

func (u *Handler) clientRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetClientsParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetClientsWithParams(r.Context(), token, realm, params)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Client](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateClient(r.Context(), token, realm, req))
	}))

	router.Route("/{clientID}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClient(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.Client](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "clientID"))
			return nil, k.UpdateClient(r.Context(), token, realm, req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteClient(r.Context(), token, realm, param(r, "clientID"))
		}))

		router.Get("/client-secret", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientSecret(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Post("/client-secret", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.RegenerateClientSecret(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Get("/service-account-user", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientServiceAccount(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Get("/user-sessions", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientUserSessions(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Get("/offline-sessions", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientOfflineSessions(r.Context(), token, realm, param(r, "clientID"))
		}))

		router.Get("/default-client-scopes", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientsDefaultScopes(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Put("/default-client-scopes/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.AddDefaultScopeToClient(r.Context(), token, realm, param(r, "clientID"), param(r, "scopeID"))
		}))
		router.Delete("/default-client-scopes/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.RemoveDefaultScopeFromClient(r.Context(), token, realm, param(r, "clientID"), param(r, "scopeID"))
		}))
		router.Get("/optional-client-scopes", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientsOptionalScopes(r.Context(), token, realm, param(r, "clientID"))
		}))
		router.Put("/optional-client-scopes/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.AddOptionalScopeToClient(r.Context(), token, realm, param(r, "clientID"), param(r, "scopeID"))
		}))
		router.Delete("/optional-client-scopes/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.RemoveOptionalScopeFromClient(r.Context(), token, realm, param(r, "clientID"), param(r, "scopeID"))
		}))

		router.Route("/scope-mappings", func(router chi.Router) {
			router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
				return k.GetClientScopeMappings(r.Context(), token, realm, param(r, "clientID"))
			}))
			u.scopeMappingRoutes(router, "clientID", scopeMappingOps{
				realmGet:        k.GetClientScopeMappingsRealmRoles,
				realmAvailable:  k.GetClientScopeMappingsRealmRolesAvailable,
				realmAdd:        k.CreateClientScopeMappingsRealmRoles,
				realmDelete:     k.DeleteClientScopeMappingsRealmRoles,
				clientGet:       k.GetClientScopeMappingsClientRoles,
				clientAvailable: k.GetClientScopeMappingsClientRolesAvailable,
				clientAdd:       k.CreateClientScopeMappingsClientRoles,
				clientDelete:    k.DeleteClientScopeMappingsClientRoles,
			})
		})

		router.Route("/roles", u.clientRoleRoutes)

		router.Post("/protocol-mappers", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ProtocolMapperRepresentation](r)
			if err != nil {
				return nil, err
			}
			return created(k.CreateClientProtocolMapper(r.Context(), token, realm, param(r, "clientID"), req))
		}))
		router.Put("/protocol-mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ProtocolMapperRepresentation](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "mapperID"))
			return nil, k.UpdateClientProtocolMapper(r.Context(), token, realm, param(r, "clientID"), param(r, "mapperID"), req)
		}))
		router.Delete("/protocol-mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteClientProtocolMapper(r.Context(), token, realm, param(r, "clientID"), param(r, "mapperID"))
		}))

		router.Route("/authz/resource-server", u.resourceServerRoutes)
	})
}

func (u *Handler) clientRoleRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetRoleParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetClientRoles(r.Context(), token, realm, param(r, "clientID"), params)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Role](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateClientRole(r.Context(), token, realm, param(r, "clientID"), req))
	}))
	router.Get("/{roleName}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetClientRole(r.Context(), token, realm, param(r, "clientID"), param(r, "roleName"))
	}))
	router.Put("/{roleName}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Role](r)
		if err != nil {
			return nil, err
		}
		req.Name = gocloak.StringP(param(r, "roleName"))
		return nil, k.UpdateRole(r.Context(), token, realm, param(r, "clientID"), req)
	}))
	router.Delete("/{roleName}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteClientRole(r.Context(), token, realm, param(r, "clientID"), param(r, "roleName"))
	}))
	router.Get("/{roleName}/users", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetUsersByRoleParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetUsersByClientRoleName(r.Context(), token, realm, param(r, "clientID"), param(r, "roleName"), params)
	}))
	router.Get("/{roleName}/groups", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetGroupsByClientRole(r.Context(), token, realm, param(r, "roleName"), param(r, "clientID"))
	}))
}

func (u *Handler) clientRegistrationRoutes(router chi.Router) {
	k := u.keycloak

	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Client](r)
		if err != nil {
			return nil, err
		}
		return createdValue(k.CreateClientRepresentation(r.Context(), token, realm, req))
	}))
	router.Get("/{clientID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetClientRepresentation(r.Context(), token, realm, param(r, "clientID"))
	}))
	router.Put("/{clientID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Client](r)
		if err != nil {
			return nil, err
		}
		req.ClientID = gocloak.StringP(param(r, "clientID"))
		return k.UpdateClientRepresentation(r.Context(), token, realm, req)
	}))
	router.Delete("/{clientID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteClientRepresentation(r.Context(), token, realm, param(r, "clientID"))
	}))
	router.Get("/{clientID}/installation", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAdapterConfiguration(r.Context(), token, realm, param(r, "clientID"))
	}))
}

type scopeMappingOps struct {
	realmGet        rolesOf
	realmAvailable  rolesOf
	realmAdd        changeRoles
	realmDelete     changeRoles
	clientGet       clientRolesOf
	clientAvailable clientRolesOf
	clientAdd       changeClientRole
	clientDelete    changeClientRole
}

func (u *Handler) scopeMappingRoutes(router chi.Router, idParam string, ops scopeMappingOps) {
	getRoles := func(fn rolesOf) http.HandlerFunc {
		return u.admin(func(r *http.Request, token, realm string) (any, error) {
			return fn(r.Context(), token, realm, param(r, idParam))
		})
	}
	getClientRoles := func(fn clientRolesOf) http.HandlerFunc {
		return u.admin(func(r *http.Request, token, realm string) (any, error) {
			return fn(r.Context(), token, realm, param(r, idParam), param(r, "selectedClientID"))
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
			return nil, fn(r.Context(), token, realm, param(r, idParam), param(r, "selectedClientID"), roles)
		})
	}

	router.Get("/realm", getRoles(ops.realmGet))
	router.Post("/realm", setRoles(ops.realmAdd))
	router.Delete("/realm", setRoles(ops.realmDelete))
	router.Get("/realm/available", getRoles(ops.realmAvailable))
	router.Get("/clients/{selectedClientID}", getClientRoles(ops.clientGet))
	router.Post("/clients/{selectedClientID}", setClientRoles(ops.clientAdd))
	router.Delete("/clients/{selectedClientID}", setClientRoles(ops.clientDelete))
	router.Get("/clients/{selectedClientID}/available", getClientRoles(ops.clientAvailable))
}

func (u *Handler) clientScopeRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetClientScopes(r.Context(), token, realm)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.ClientScope](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateClientScope(r.Context(), token, realm, req))
	}))

	router.Route("/{scopeID}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientScope(r.Context(), token, realm, param(r, "scopeID"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ClientScope](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "scopeID"))
			return nil, k.UpdateClientScope(r.Context(), token, realm, req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteClientScope(r.Context(), token, realm, param(r, "scopeID"))
		}))

		router.Get("/protocol-mappers", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientScopeProtocolMappers(r.Context(), token, realm, param(r, "scopeID"))
		}))
		router.Post("/protocol-mappers", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ProtocolMappers](r)
			if err != nil {
				return nil, err
			}
			return created(k.CreateClientScopeProtocolMapper(r.Context(), token, realm, param(r, "scopeID"), req))
		}))
		router.Get("/protocol-mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetClientScopeProtocolMapper(r.Context(), token, realm, param(r, "scopeID"), param(r, "mapperID"))
		}))
		router.Put("/protocol-mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ProtocolMappers](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "mapperID"))
			return nil, k.UpdateClientScopeProtocolMapper(r.Context(), token, realm, param(r, "scopeID"), req)
		}))
		router.Delete("/protocol-mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteClientScopeProtocolMapper(r.Context(), token, realm, param(r, "scopeID"), param(r, "mapperID"))
		}))

		router.Route("/scope-mappings", func(router chi.Router) {
			u.scopeMappingRoutes(router, "scopeID", scopeMappingOps{
				realmGet:        k.GetClientScopesScopeMappingsRealmRoles,
				realmAvailable:  k.GetClientScopesScopeMappingsRealmRolesAvailable,
				realmAdd:        k.CreateClientScopesScopeMappingsRealmRoles,
				realmDelete:     k.DeleteClientScopesScopeMappingsRealmRoles,
				clientGet:       k.GetClientScopesScopeMappingsClientRoles,
				clientAvailable: k.GetClientScopesScopeMappingsClientRolesAvailable,
				clientAdd:       k.CreateClientScopesScopeMappingsClientRoles,
				clientDelete:    k.DeleteClientScopesScopeMappingsClientRoles,
			})
		})
	})
}

func (u *Handler) resourceServerRoutes(router chi.Router) {
	k := u.keycloak
	client := func(r *http.Request) string { return param(r, "clientID") }

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetResourceServer(r.Context(), token, realm, client(r))
	}))

	router.Get("/resource", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetResourceParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetResources(r.Context(), token, realm, client(r), params)
	}))
	router.Post("/resource", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.ResourceRepresentation](r)
		if err != nil {
			return nil, err
		}
		return createdValue(k.CreateResource(r.Context(), token, realm, client(r), req))
	}))
	router.Get("/resource/{resourceID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetResource(r.Context(), token, realm, client(r), param(r, "resourceID"))
	}))
	router.Put("/resource/{resourceID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.ResourceRepresentation](r)
		if err != nil {
			return nil, err
		}
		req.ID = gocloak.StringP(param(r, "resourceID"))
		return nil, k.UpdateResource(r.Context(), token, realm, client(r), req)
	}))
	router.Delete("/resource/{resourceID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteResource(r.Context(), token, realm, client(r), param(r, "resourceID"))
	}))

	router.Get("/scope", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetScopeParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetScopes(r.Context(), token, realm, client(r), params)
	}))
	router.Post("/scope", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.ScopeRepresentation](r)
		if err != nil {
			return nil, err
		}
		return createdValue(k.CreateScope(r.Context(), token, realm, client(r), req))
	}))
	router.Get("/scope/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetScope(r.Context(), token, realm, client(r), param(r, "scopeID"))
	}))
	router.Put("/scope/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.ScopeRepresentation](r)
		if err != nil {
			return nil, err
		}
		req.ID = gocloak.StringP(param(r, "scopeID"))
		return nil, k.UpdateScope(r.Context(), token, realm, client(r), req)
	}))
	router.Delete("/scope/{scopeID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteScope(r.Context(), token, realm, client(r), param(r, "scopeID"))
	}))

	router.Get("/policy", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetPolicyParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetPolicies(r.Context(), token, realm, client(r), params)
	}))
	router.Post("/policy", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.PolicyRepresentation](r)
		if err != nil {
			return nil, err
		}
		return createdValue(k.CreatePolicy(r.Context(), token, realm, client(r), req))
	}))
	router.Get("/policy/{policyID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetPolicy(r.Context(), token, realm, client(r), param(r, "policyID"))
	}))
	router.Put("/policy/{policyID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.PolicyRepresentation](r)
		if err != nil {
			return nil, err
		}
		req.ID = gocloak.StringP(param(r, "policyID"))
		return nil, k.UpdatePolicy(r.Context(), token, realm, client(r), req)
	}))
	router.Delete("/policy/{policyID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeletePolicy(r.Context(), token, realm, client(r), param(r, "policyID"))
	}))
	router.Get("/policy/{policyID}/associatedPolicies", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAuthorizationPolicyAssociatedPolicies(r.Context(), token, realm, client(r), param(r, "policyID"))
	}))
	router.Get("/policy/{policyID}/resources", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAuthorizationPolicyResources(r.Context(), token, realm, client(r), param(r, "policyID"))
	}))
	router.Get("/policy/{policyID}/scopes", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAuthorizationPolicyScopes(r.Context(), token, realm, client(r), param(r, "policyID"))
	}))
	router.Get("/policy/{policyID}/dependentPolicies", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetDependentPermissions(r.Context(), token, realm, client(r), param(r, "policyID"))
	}))

	router.Get("/permission", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetPermissionParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetPermissions(r.Context(), token, realm, client(r), params)
	}))
	router.Post("/permission", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.PermissionRepresentation](r)
		if err != nil {
			return nil, err
		}
		return createdValue(k.CreatePermission(r.Context(), token, realm, client(r), req))
	}))
	router.Get("/permission/{permissionID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetPermission(r.Context(), token, realm, client(r), param(r, "permissionID"))
	}))
	router.Put("/permission/{permissionID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.PermissionRepresentation](r)
		if err != nil {
			return nil, err
		}
		req.ID = gocloak.StringP(param(r, "permissionID"))
		return nil, k.UpdatePermission(r.Context(), token, realm, client(r), req)
	}))
	router.Delete("/permission/{permissionID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeletePermission(r.Context(), token, realm, client(r), param(r, "permissionID"))
	}))
	router.Get("/permission/{permissionID}/resources", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetPermissionResources(r.Context(), token, realm, client(r), param(r, "permissionID"))
	}))
	router.Get("/permission/{permissionID}/scopes", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetPermissionScopes(r.Context(), token, realm, client(r), param(r, "permissionID"))
	}))
}
