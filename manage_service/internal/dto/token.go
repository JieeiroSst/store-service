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

type LoginAdmin struct {
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

type Login struct {
	ClientID     string
	ClientSecret string
	Realm        string
	Username     string
	Password     string
}

type LoginOTP struct {
	ClientID     string
	ClientSecret string
	Realm        string
	Username     string
	Password     string
	OTP          string
}

type LoginClient struct {
	ClientID     string
	ClientSecret string
	Realm        string
}

type RefreshToken struct {
	RefreshToken string
	ClientID     string
	ClientSecret string
	Realm        string
}

type SetPassword struct {
	Token     string
	UserID    string
	Realm     string
	Password  string
	Temporary bool
}

type Logout struct {
	ClientID     string
	ClientSecret string
	Realm        string
	RefreshToken string
}

type UserInfo struct {
	AccessToken string
	Realm       string
}
