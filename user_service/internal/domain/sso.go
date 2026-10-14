package domain

type ExternalIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	Username      string
	AvatarURL     string
	Scopes        []string
	Profile       []byte
}
