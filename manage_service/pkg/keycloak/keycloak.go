package keycloak

import (
	keycloak "github.com/Nerzal/gocloak/v13"
)

func NewKeycloak(host string) *keycloak.GoCloak {
	client := keycloak.NewClient(host)
	return client
}
