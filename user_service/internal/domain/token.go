package domain

type AccessClaims struct {
	UserID   int
	Username string
}

type Session struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}
