package model

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

type Session struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}
