package http

import "github.com/Nerzal/gocloak/v13"

type TokenRequest struct {
	gocloak.TokenOptions
	ClientSecret  *string   `json:"client_secret,omitempty"`
	Scopes        *[]string `json:"scopes,omitempty"`
	ResponseTypes *[]string `json:"response_types,omitempty"`
}

type IntrospectRequest struct {
	Token        string `json:"token"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type RevokeTokenRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RefreshToken string `json:"refreshToken"`
}

type LogoutPublicClientRequest struct {
	ClientID     string `json:"clientId"`
	RefreshToken string `json:"refreshToken"`
}

type TokenExchangeRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	Token        string `json:"token"`
	TargetClient string `json:"targetClient"`
	UserID       string `json:"userId"`
}

type RequestingPartyTokenRequest struct {
	gocloak.RequestingPartyTokenOptions
	Permissions *[]string `json:"permissions,omitempty"`
}

type ResetPasswordRequest struct {
	Password  string `json:"password"`
	Temporary bool   `json:"temporary"`
}

type ExecuteActionsEmailRequest struct {
	Actions     []string `json:"actions"`
	ClientID    *string  `json:"clientId,omitempty"`
	Lifespan    *int     `json:"lifespan,omitempty"`
	RedirectURI *string  `json:"redirectUri,omitempty"`
}

type CredentialLabelRequest struct {
	UserLabel string `json:"userLabel"`
}

type ImportIdentityProviderRequest struct {
	FromURL    string `json:"fromUrl"`
	ProviderID string `json:"providerId"`
}

type CreatedResponse struct {
	ID string `json:"id"`
}
