package http

import (
	"errors"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
)

func (u *Handler) keycloakRoutes(router chi.Router) {
	router.Route("/realms/{realm}", u.openIDRoutes)
	router.Route("/admin", u.adminRoutes)
}

func (u *Handler) adminRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/server-info", u.admin(func(r *http.Request, token, _ string) (any, error) {
		return k.GetServerInfo(r.Context(), token)
	}))

	router.Get("/realms", u.admin(func(r *http.Request, token, _ string) (any, error) {
		return k.GetRealms(r.Context(), token)
	}))
	router.Post("/realms", u.admin(func(r *http.Request, token, _ string) (any, error) {
		req, err := body[gocloak.RealmRepresentation](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateRealm(r.Context(), token, req))
	}))

	router.Route("/realms/{realm}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetRealm(r.Context(), token, realm)
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.RealmRepresentation](r)
			if err != nil {
				return nil, err
			}
			if req.Realm == nil {
				req.Realm = gocloak.StringP(realm)
			}
			return nil, k.UpdateRealm(r.Context(), token, req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteRealm(r.Context(), token, realm)
		}))
		router.Post("/clear-realm-cache", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.ClearRealmCache(r.Context(), token, realm)
		}))
		router.Post("/clear-user-cache", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.ClearUserCache(r.Context(), token, realm)
		}))
		router.Post("/clear-keys-cache", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.ClearKeysCache(r.Context(), token, realm)
		}))
		router.Get("/keys", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetKeyStoreConfig(r.Context(), token, realm)
		}))
		router.Get("/events", u.admin(func(r *http.Request, token, realm string) (any, error) {
			r.URL.RawQuery = withoutKey(r.URL.Query(), "type")
			params, err := query[gocloak.GetEventsParams](r)
			if err != nil {
				return nil, err
			}
			return k.GetEvents(r.Context(), token, realm, params)
		}))
		router.Get("/credential-registrators", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetCredentialRegistrators(r.Context(), token, realm)
		}))
		router.Delete("/sessions/{sessionID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.LogoutUserSession(r.Context(), token, realm, param(r, "sessionID"))
		}))

		router.Route("/users", u.userRoutes)
		router.Route("/groups", u.groupRoutes)
		router.Get("/group-by-path/*", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetGroupByPath(r.Context(), token, realm, param(r, "*"))
		}))
		router.Route("/default-groups", u.defaultGroupRoutes)
		router.Route("/roles", u.realmRoleRoutes)
		router.Route("/roles-by-id", u.roleByIDRoutes)
		router.Route("/clients", u.clientRoutes)
		router.Route("/clients-registrations", u.clientRegistrationRoutes)
		router.Route("/client-scopes", u.clientScopeRoutes)
		router.Get("/default-default-client-scopes", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetDefaultDefaultClientScopes(r.Context(), token, realm)
		}))
		router.Get("/default-optional-client-scopes", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetDefaultOptionalClientScopes(r.Context(), token, realm)
		}))
		router.Route("/components", u.componentRoutes)
		router.Route("/authentication", u.authenticationRoutes)
		router.Route("/identity-providers", u.identityProviderRoutes)
	})
}

func (u *Handler) componentRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetComponentsParams](r)
		if err != nil {
			return nil, err
		}
		if params == (gocloak.GetComponentsParams{}) {
			return k.GetComponents(r.Context(), token, realm)
		}
		return k.GetComponentsWithParams(r.Context(), token, realm, params)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Component](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateComponent(r.Context(), token, realm, req))
	}))
	router.Get("/{componentID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetComponent(r.Context(), token, realm, param(r, "componentID"))
	}))
	router.Put("/{componentID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Component](r)
		if err != nil {
			return nil, err
		}
		req.ID = gocloak.StringP(param(r, "componentID"))
		return nil, k.UpdateComponent(r.Context(), token, realm, req)
	}))
	router.Delete("/{componentID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteComponent(r.Context(), token, realm, param(r, "componentID"))
	}))
}

func (u *Handler) authenticationRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/flows", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAuthenticationFlows(r.Context(), token, realm)
	}))
	router.Post("/flows", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.AuthenticationFlowRepresentation](r)
		if err != nil {
			return nil, err
		}
		if err := k.CreateAuthenticationFlow(r.Context(), token, realm, req); err != nil {
			return nil, err
		}
		return createdResult{req}, nil
	}))
	router.Get("/flows/{flow}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAuthenticationFlow(r.Context(), token, realm, param(r, "flow"))
	}))
	router.Put("/flows/{flow}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.AuthenticationFlowRepresentation](r)
		if err != nil {
			return nil, err
		}
		return k.UpdateAuthenticationFlow(r.Context(), token, realm, req, param(r, "flow"))
	}))
	router.Delete("/flows/{flow}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteAuthenticationFlow(r.Context(), token, realm, param(r, "flow"))
	}))
	router.Get("/flows/{flow}/executions", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetAuthenticationExecutions(r.Context(), token, realm, param(r, "flow"))
	}))
	router.Put("/flows/{flow}/executions", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.ModifyAuthenticationExecutionRepresentation](r)
		if err != nil {
			return nil, err
		}
		return nil, k.UpdateAuthenticationExecution(r.Context(), token, realm, param(r, "flow"), req)
	}))
	router.Post("/flows/{flow}/executions/execution", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.CreateAuthenticationExecutionRepresentation](r)
		if err != nil {
			return nil, err
		}
		if err := k.CreateAuthenticationExecution(r.Context(), token, realm, param(r, "flow"), req); err != nil {
			return nil, err
		}
		return createdResult{req}, nil
	}))
	router.Post("/flows/{flow}/executions/flow", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.CreateAuthenticationExecutionFlowRepresentation](r)
		if err != nil {
			return nil, err
		}
		if err := k.CreateAuthenticationExecutionFlow(r.Context(), token, realm, param(r, "flow"), req); err != nil {
			return nil, err
		}
		return createdResult{req}, nil
	}))
	router.Delete("/executions/{executionID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteAuthenticationExecution(r.Context(), token, realm, param(r, "executionID"))
	}))

	router.Get("/required-actions", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetRequiredActions(r.Context(), token, realm)
	}))
	router.Post("/register-required-action", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.RequiredActionProviderRepresentation](r)
		if err != nil {
			return nil, err
		}
		if err := k.RegisterRequiredAction(r.Context(), token, realm, req); err != nil {
			return nil, err
		}
		return createdResult{req}, nil
	}))
	router.Get("/required-actions/{alias}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetRequiredAction(r.Context(), token, realm, param(r, "alias"))
	}))
	router.Put("/required-actions/{alias}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.RequiredActionProviderRepresentation](r)
		if err != nil {
			return nil, err
		}
		req.Alias = gocloak.StringP(param(r, "alias"))
		return nil, k.UpdateRequiredAction(r.Context(), token, realm, req)
	}))
	router.Delete("/required-actions/{alias}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.DeleteRequiredAction(r.Context(), token, realm, param(r, "alias"))
	}))
}

func (u *Handler) identityProviderRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetIdentityProviders(r.Context(), token, realm)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.IdentityProviderRepresentation](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateIdentityProvider(r.Context(), token, realm, req))
	}))
	router.Post("/import-config", u.admin(func(r *http.Request, token, realm string) (any, error) {
		if file, header, err := r.FormFile("file"); err == nil {
			defer file.Close()
			providerID := r.FormValue("providerId")
			if providerID == "" {
				return nil, badRequestError{errors.New("providerId is required")}
			}
			return k.ImportIdentityProviderConfigFromFile(r.Context(), token, realm, providerID, header.Filename, file)
		}
		req, err := body[ImportIdentityProviderRequest](r)
		if err != nil {
			return nil, err
		}
		return k.ImportIdentityProviderConfig(r.Context(), token, realm, req.FromURL, req.ProviderID)
	}))

	router.Route("/{alias}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetIdentityProvider(r.Context(), token, realm, param(r, "alias"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.IdentityProviderRepresentation](r)
			if err != nil {
				return nil, err
			}
			return nil, k.UpdateIdentityProvider(r.Context(), token, realm, param(r, "alias"), req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteIdentityProvider(r.Context(), token, realm, param(r, "alias"))
		}))
		router.Get("/export", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.ExportIDPPublicBrokerConfig(r.Context(), token, realm, param(r, "alias"))
		}))
		router.Get("/mappers", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetIdentityProviderMappers(r.Context(), token, realm, param(r, "alias"))
		}))
		router.Post("/mappers", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.IdentityProviderMapper](r)
			if err != nil {
				return nil, err
			}
			return created(k.CreateIdentityProviderMapper(r.Context(), token, realm, param(r, "alias"), req))
		}))
		router.Get("/mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetIdentityProviderMapperByID(r.Context(), token, realm, param(r, "alias"), param(r, "mapperID"))
		}))
		router.Put("/mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.IdentityProviderMapper](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "mapperID"))
			return nil, k.UpdateIdentityProviderMapper(r.Context(), token, realm, param(r, "alias"), req)
		}))
		router.Delete("/mappers/{mapperID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteIdentityProviderMapper(r.Context(), token, realm, param(r, "alias"), param(r, "mapperID"))
		}))
	})
}
