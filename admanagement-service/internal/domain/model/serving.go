package model

type TargetContext struct {
	UserID      *uint  `json:"user_id,omitempty"`
	SessionID   string `json:"session_id,omitempty"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	ReferrerURL string `json:"referrer_url,omitempty"`
	PageURL     string `json:"page_url,omitempty"`
	Country     string `json:"country,omitempty"`
	Device      string `json:"device,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	HourOfDay   int    `json:"hour_of_day,omitempty"`
}

type ServedAd struct {
	Ad           Ad   `json:"ad"`
	ImpressionID uint `json:"impression_id"`
}
