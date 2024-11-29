package user

type UserResponse struct {
	Status   int      `json:"status,omitempty"`
	Message  string   `json:"message,omitempty"`
	UserInfo UserInfo `json:"user_info,omitempty"`
}

type UserInfo struct {
	UserID       int    `json:"user_id,omitempty"`
	CountryID    string `json:"country_id,omitempty"`
	CountryName  string `json:"country_name,omitempty"`
	Email        string `json:"email,omitempty"`
	Address      string `json:"address,omitempty"`
	CustomerName string `json:"customer_name,omitempty"`
	Phone        string `json:"phone,omitempty"`
}
