package model

type Author struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url,omitempty"`
}
