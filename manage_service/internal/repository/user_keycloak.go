package repository

import (
	keycloak "github.com/Nerzal/gocloak/v13"
)

type UserKeycloak interface {
}

type UserKeycloakRepo struct {
	client keycloak.GoCloak
}

func NewUserKeycloakRepo(client keycloak.GoCloak) UserKeycloak {
	return &UserKeycloakRepo{
		client: client,
	}
}
