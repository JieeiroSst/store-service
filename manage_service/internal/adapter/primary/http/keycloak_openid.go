package http

import (
	"net/http"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
)

func (u *Handler) openIDRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/issuer", u.handle(func(r *http.Request) (any, error) {
		return k.GetIssuer(r.Context(), param(r, "realm"))
	}))
	router.Get("/protocol/certs", u.handle(func(r *http.Request) (any, error) {
		return k.GetCerts(r.Context(), param(r, "realm"))
	}))
	router.Post("/protocol/token", u.handle(func(r *http.Request) (any, error) {
		req, err := body[TokenRequest](r)
		if err != nil {
			return nil, err
		}
		options := req.TokenOptions
		options.ClientSecret = req.ClientSecret
		options.Scopes = req.Scopes
		options.ResponseTypes = req.ResponseTypes
		return k.GetToken(r.Context(), param(r, "realm"), options)
	}))
	router.Post("/protocol/introspect", u.handle(func(r *http.Request) (any, error) {
		req, err := body[IntrospectRequest](r)
		if err != nil {
			return nil, err
		}
		return k.RetrospectToken(r.Context(), req.Token, req.ClientID, req.ClientSecret, param(r, "realm"))
	}))
	router.Post("/protocol/revoke", u.handle(func(r *http.Request) (any, error) {
		req, err := body[RevokeTokenRequest](r)
		if err != nil {
			return nil, err
		}
		return nil, k.RevokeToken(r.Context(), param(r, "realm"), req.ClientID, req.ClientSecret, req.RefreshToken)
	}))
	router.Post("/protocol/token-exchange", u.handle(func(r *http.Request) (any, error) {
		req, err := body[TokenExchangeRequest](r)
		if err != nil {
			return nil, err
		}
		return k.LoginClientTokenExchange(r.Context(), req.ClientID, req.Token, req.ClientSecret, param(r, "realm"), req.TargetClient, req.UserID)
	}))
	router.Post("/protocol/logout-public", u.user(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[LogoutPublicClientRequest](r)
		if err != nil {
			return nil, err
		}
		return nil, k.LogoutPublicClient(r.Context(), req.ClientID, realm, token, req.RefreshToken)
	}))
	router.Get("/protocol/userinfo", u.user(func(r *http.Request, token, realm string) (any, error) {
		return k.GetRawUserInfo(r.Context(), token, realm)
	}))
	router.Get("/protocol/decode", u.user(func(r *http.Request, token, realm string) (any, error) {
		return k.DecodeAccessToken(r.Context(), token, realm)
	}))

	router.Route("/authz", func(router chi.Router) {
		router.Post("/rpt", u.user(func(r *http.Request, token, realm string) (any, error) {
			options, err := rptOptions(r)
			if err != nil {
				return nil, err
			}
			return k.GetRequestingPartyToken(r.Context(), token, realm, options)
		}))
		router.Post("/rpt/permissions", u.user(func(r *http.Request, token, realm string) (any, error) {
			options, err := rptOptions(r)
			if err != nil {
				return nil, err
			}
			return k.GetRequestingPartyPermissions(r.Context(), token, realm, options)
		}))
		router.Post("/rpt/decision", u.user(func(r *http.Request, token, realm string) (any, error) {
			options, err := rptOptions(r)
			if err != nil {
				return nil, err
			}
			return k.GetRequestingPartyPermissionDecision(r.Context(), token, realm, options)
		}))

		router.Post("/permission-tickets", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[[]gocloak.CreatePermissionTicketParams](r)
			if err != nil {
				return nil, err
			}
			return k.CreatePermissionTicket(r.Context(), token, realm, req)
		}))

		router.Get("/uma-permissions", u.user(func(r *http.Request, token, realm string) (any, error) {
			q := r.URL.Query()
			params := gocloak.GetUserPermissionParams{
				ScopeID:     optString(q.Get("scopeId")),
				ResourceID:  optString(q.Get("resourceId")),
				Owner:       optString(q.Get("owner")),
				Requester:   optString(q.Get("requester")),
				ReturnNames: optString(q.Get("returnNames")),
			}
			if v, err := strconv.ParseBool(q.Get("granted")); err == nil {
				params.Granted = &v
			}
			if v, err := strconv.Atoi(q.Get("first")); err == nil {
				params.First = &v
			}
			if v, err := strconv.Atoi(q.Get("max")); err == nil {
				params.Max = &v
			}
			return k.GetUserPermissions(r.Context(), token, realm, params)
		}))
		router.Post("/uma-permissions", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.PermissionGrantParams](r)
			if err != nil {
				return nil, err
			}
			return createdValue(k.GrantUserPermission(r.Context(), token, realm, req))
		}))
		router.Put("/uma-permissions", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.PermissionGrantParams](r)
			if err != nil {
				return nil, err
			}
			return k.UpdateUserPermission(r.Context(), token, realm, req)
		}))
		router.Delete("/uma-permissions/{ticketID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteUserPermission(r.Context(), token, realm, param(r, "ticketID"))
		}))

		router.Get("/resources", u.user(func(r *http.Request, token, realm string) (any, error) {
			params, err := query[gocloak.GetResourceParams](r)
			if err != nil {
				return nil, err
			}
			return k.GetResourcesClient(r.Context(), token, realm, params)
		}))
		router.Post("/resources", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ResourceRepresentation](r)
			if err != nil {
				return nil, err
			}
			return createdValue(k.CreateResourceClient(r.Context(), token, realm, req))
		}))
		router.Get("/resources/{resourceID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			return k.GetResourceClient(r.Context(), token, realm, param(r, "resourceID"))
		}))
		router.Put("/resources/{resourceID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ResourceRepresentation](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "resourceID"))
			return nil, k.UpdateResourceClient(r.Context(), token, realm, req)
		}))
		router.Delete("/resources/{resourceID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteResourceClient(r.Context(), token, realm, param(r, "resourceID"))
		}))
		router.Post("/resources/{resourceID}/policies", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ResourcePolicyRepresentation](r)
			if err != nil {
				return nil, err
			}
			return createdValue(k.CreateResourcePolicy(r.Context(), token, realm, param(r, "resourceID"), req))
		}))

		router.Get("/resource-policies", u.user(func(r *http.Request, token, realm string) (any, error) {
			params, err := query[gocloak.GetResourcePoliciesParams](r)
			if err != nil {
				return nil, err
			}
			return k.GetResourcePolicies(r.Context(), token, realm, params)
		}))
		router.Get("/resource-policies/{permissionID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			return k.GetResourcePolicy(r.Context(), token, realm, param(r, "permissionID"))
		}))
		router.Put("/resource-policies/{permissionID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.ResourcePolicyRepresentation](r)
			if err != nil {
				return nil, err
			}
			return nil, k.UpdateResourcePolicy(r.Context(), token, realm, param(r, "permissionID"), req)
		}))
		router.Delete("/resource-policies/{permissionID}", u.user(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteResourcePolicy(r.Context(), token, realm, param(r, "permissionID"))
		}))
	})
}

func rptOptions(r *http.Request) (gocloak.RequestingPartyTokenOptions, error) {
	req, err := body[RequestingPartyTokenRequest](r)
	if err != nil {
		return gocloak.RequestingPartyTokenOptions{}, err
	}
	options := req.RequestingPartyTokenOptions
	options.Permissions = req.Permissions
	return options, nil
}

func optString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
