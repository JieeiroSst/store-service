package domain

const DefaultRole = "user"

type UserRoles struct {
	Roles          []string
	EffectiveRoles []string
	PrimaryRole    string
}
