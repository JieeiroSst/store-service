package http

import (
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/go-chi/chi/v5"
)

func (u *Handler) userRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetUsersParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetUsers(r.Context(), token, realm, params)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.User](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateUserRepresentation(r.Context(), token, realm, req))
	}))
	router.Get("/count", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetUsersParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetUserCount(r.Context(), token, realm, params)
	}))

	router.Route("/{userID}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetUserByID(r.Context(), token, realm, param(r, "userID"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.User](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "userID"))
			return nil, k.UpdateUser(r.Context(), token, realm, req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteUser(r.Context(), token, realm, param(r, "userID"))
		}))

		router.Put("/reset-password", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[ResetPasswordRequest](r)
			if err != nil {
				return nil, err
			}
			return nil, k.SetPassword(r.Context(), token, param(r, "userID"), realm, req.Password, req.Temporary)
		}))
		router.Put("/execute-actions-email", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[ExecuteActionsEmailRequest](r)
			if err != nil {
				return nil, err
			}
			return nil, k.ExecuteActionsEmail(r.Context(), token, realm, gocloak.ExecuteActionsEmail{
				UserID:      gocloak.StringP(param(r, "userID")),
				Actions:     &req.Actions,
				ClientID:    req.ClientID,
				Lifespan:    req.Lifespan,
				RedirectURI: req.RedirectURI,
			})
		}))
		router.Put("/send-verify-email", u.admin(func(r *http.Request, token, realm string) (any, error) {
			q := r.URL.Query()
			return nil, k.SendVerifyEmail(r.Context(), token, param(r, "userID"), realm, gocloak.SendVerificationMailParams{
				ClientID:    optString(q.Get("client_id")),
				RedirectURI: optString(q.Get("redirect_uri")),
			})
		}))
		router.Get("/brute-force", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetUserBruteForceDetectionStatus(r.Context(), token, realm, param(r, "userID"))
		}))
		router.Delete("/consents/{clientID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.RevokeUserConsents(r.Context(), token, realm, param(r, "userID"), param(r, "clientID"))
		}))

		router.Get("/sessions", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetUserSessions(r.Context(), token, realm, param(r, "userID"))
		}))
		router.Get("/offline-sessions/{clientID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetUserOfflineSessionsForClient(r.Context(), token, realm, param(r, "userID"), param(r, "clientID"))
		}))
		router.Post("/logout", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.LogoutAllSessions(r.Context(), token, realm, param(r, "userID"))
		}))

		router.Get("/groups", u.admin(func(r *http.Request, token, realm string) (any, error) {
			params, err := query[gocloak.GetGroupsParams](r)
			if err != nil {
				return nil, err
			}
			return k.GetUserGroups(r.Context(), token, realm, param(r, "userID"), params)
		}))
		router.Put("/groups/{groupID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.AddUserToGroup(r.Context(), token, realm, param(r, "userID"), param(r, "groupID"))
		}))
		router.Delete("/groups/{groupID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteUserFromGroup(r.Context(), token, realm, param(r, "userID"), param(r, "groupID"))
		}))

		router.Get("/credentials", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetCredentials(r.Context(), token, realm, param(r, "userID"))
		}))
		router.Delete("/credentials/{credentialID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteCredentials(r.Context(), token, realm, param(r, "userID"), param(r, "credentialID"))
		}))
		router.Put("/credentials/{credentialID}/userLabel", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[CredentialLabelRequest](r)
			if err != nil {
				return nil, err
			}
			return nil, k.UpdateCredentialUserLabel(r.Context(), token, realm, param(r, "userID"), param(r, "credentialID"), req.UserLabel)
		}))
		router.Post("/credentials/{credentialID}/moveToFirst", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.MoveCredentialToFirst(r.Context(), token, realm, param(r, "userID"), param(r, "credentialID"))
		}))
		router.Post("/credentials/{credentialID}/moveAfter/{previousCredentialID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.MoveCredentialBehind(r.Context(), token, realm, param(r, "userID"), param(r, "credentialID"), param(r, "previousCredentialID"))
		}))
		router.Get("/configured-user-storage-credential-types", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetConfiguredUserStorageCredentialTypes(r.Context(), token, realm, param(r, "userID"))
		}))
		router.Put("/disable-credential-types", u.admin(func(r *http.Request, token, realm string) (any, error) {
			types, err := body[[]string](r)
			if err != nil {
				return nil, err
			}
			return nil, k.DisableAllCredentialsByType(r.Context(), token, realm, param(r, "userID"), types)
		}))

		router.Get("/federated-identity", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetUserFederatedIdentities(r.Context(), token, realm, param(r, "userID"))
		}))
		router.Post("/federated-identity/{providerID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.FederatedIdentityRepresentation](r)
			if err != nil {
				return nil, err
			}
			if err := k.CreateUserFederatedIdentity(r.Context(), token, realm, param(r, "userID"), param(r, "providerID"), req); err != nil {
				return nil, err
			}
			return createdResult{req}, nil
		}))
		router.Delete("/federated-identity/{providerID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteUserFederatedIdentity(r.Context(), token, realm, param(r, "userID"), param(r, "providerID"))
		}))

		router.Route("/role-mappings", func(router chi.Router) {
			u.roleMappingRoutes(router, "userID", roleMappingOps{
				all:             k.GetRoleMappingByUserID,
				realmGet:        k.GetRealmRolesByUserID,
				realmComposite:  k.GetCompositeRealmRolesByUserID,
				realmAvailable:  k.GetAvailableRealmRolesByUserID,
				realmAdd:        k.AddRealmRoleToUser,
				realmDelete:     k.DeleteRealmRoleFromUser,
				clientGet:       k.GetClientRolesByUserID,
				clientComposite: k.GetCompositeClientRolesByUserID,
				clientAvailable: k.GetAvailableClientRolesByUserID,
				clientAdd:       k.AddClientRolesToUser,
				clientDelete:    k.DeleteClientRolesFromUser,
			})
		})
	})
}

func (u *Handler) groupRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetGroupsParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetGroups(r.Context(), token, realm, params)
	}))
	router.Post("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		req, err := body[gocloak.Group](r)
		if err != nil {
			return nil, err
		}
		return created(k.CreateGroup(r.Context(), token, realm, req))
	}))
	router.Get("/count", u.admin(func(r *http.Request, token, realm string) (any, error) {
		params, err := query[gocloak.GetGroupsParams](r)
		if err != nil {
			return nil, err
		}
		return k.GetGroupsCount(r.Context(), token, realm, params)
	}))

	router.Route("/{groupID}", func(router chi.Router) {
		router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return k.GetGroup(r.Context(), token, realm, param(r, "groupID"))
		}))
		router.Put("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.Group](r)
			if err != nil {
				return nil, err
			}
			req.ID = gocloak.StringP(param(r, "groupID"))
			return nil, k.UpdateGroup(r.Context(), token, realm, req)
		}))
		router.Delete("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
			return nil, k.DeleteGroup(r.Context(), token, realm, param(r, "groupID"))
		}))
		router.Post("/children", u.admin(func(r *http.Request, token, realm string) (any, error) {
			req, err := body[gocloak.Group](r)
			if err != nil {
				return nil, err
			}
			return created(k.CreateChildGroup(r.Context(), token, realm, param(r, "groupID"), req))
		}))
		router.Get("/members", u.admin(func(r *http.Request, token, realm string) (any, error) {
			params, err := query[gocloak.GetGroupsParams](r)
			if err != nil {
				return nil, err
			}
			return k.GetGroupMembers(r.Context(), token, realm, param(r, "groupID"), params)
		}))

		router.Route("/role-mappings", func(router chi.Router) {
			u.roleMappingRoutes(router, "groupID", roleMappingOps{
				all:             k.GetRoleMappingByGroupID,
				realmGet:        k.GetRealmRolesByGroupID,
				realmComposite:  k.GetCompositeRealmRolesByGroupID,
				realmAvailable:  k.GetAvailableRealmRolesByGroupID,
				realmAdd:        k.AddRealmRoleToGroup,
				realmDelete:     k.DeleteRealmRoleFromGroup,
				clientGet:       k.GetClientRolesByGroupID,
				clientComposite: k.GetCompositeClientRolesByGroupID,
				clientAvailable: k.GetAvailableClientRolesByGroupID,
				clientAdd:       k.AddClientRolesToGroup,
				clientDelete:    k.DeleteClientRoleFromGroup,
			})
		})
	})
}

func (u *Handler) defaultGroupRoutes(router chi.Router) {
	k := u.keycloak

	router.Get("/", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return k.GetDefaultGroups(r.Context(), token, realm)
	}))
	router.Put("/{groupID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.AddDefaultGroup(r.Context(), token, realm, param(r, "groupID"))
	}))
	router.Delete("/{groupID}", u.admin(func(r *http.Request, token, realm string) (any, error) {
		return nil, k.RemoveDefaultGroup(r.Context(), token, realm, param(r, "groupID"))
	}))
}
