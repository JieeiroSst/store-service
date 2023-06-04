package repository

import (
	keycloak "github.com/Nerzal/gocloak/v13"
)

type Repositories struct {
	UserKeycloak
}

func NewRepositories(client *keycloak.GoCloak) *Repositories {
	return &Repositories{
		UserKeycloak: NewUserKeycloakRepo(client),
	}
}
