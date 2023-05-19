package repository

import (
	keycloak "github.com/Nerzal/gocloak/v13"
	"gorm.io/gorm"
)

type Repositories struct {
	UserKeycloak
	UserDB
}

func NewRepositories(db *gorm.DB, client keycloak.GoCloak) *Repositories {
	return &Repositories{
		UserKeycloak: NewUserKeycloakRepo(client),
		UserDB:       NewUserDBRepo(db),
	}
}
