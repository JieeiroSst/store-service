package dto

type Track struct {
	ID      string `json:"id"`
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Message string `json:"message"`
}
