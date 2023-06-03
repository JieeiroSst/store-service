package dto

type Token struct {
	Token string
}

type CreateUser struct {
	Token     string
	Realm     string
	FirstName string
	LastName  string
	Email     string
	Enabled   bool
	Username  string
}

type IntrospectToken struct {
	Token        string
	ClientID     string
	ClientSecret string
	Realm        string
}

type Login struct {
	User      string
	Password  string
	RealmName string
}

type Client struct {
	Token      string
	Realm      string
	ClientName string
}

type TokenInfo struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresIn        int
	RefreshExpiresIn int
	Scope            string
}
