package dto

import (
	"mime/multipart"
	"time"
)

type Subscription struct {
	SubscriptionID int    `json:"subscription_id"`
	Name           string `json:"name"`
	SubscribedFrom string `json:"subscribed_from"`
	ValidUpto      bool   `json:"valid_upto"`
}

type Tag struct {
	TagID   int    `json:"tag_id"`
	VideoID int    `json:"video_id"`
	Value   string `json:"value"`
}

type Video struct {
	VideoID      int       `json:"video_id"`
	Description  string    `json:"description"`
	ThumbnailURL string    `json:"thumbnail_url"`
	StreamURL    string    `json:"stream_url"`
	TagID        int       `json:"tag_id"`
	UploadedOn   time.Time `json:"uploaded_on"`
	Tag          Tag       `json:"tag"`
}

type View struct {
	ViewID    int       `json:"view_id"`
	UserID    int       `json:"user_id"`
	VideoID   int       `json:"video_id"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
	TotalView int       `json:"total_view"`
}

type UploadVideoRequest struct {
	FileHeader *multipart.FileHeader `json:"_"`
	Video      Video                 `json:"video"`
	Tag        Tag                   `json:"tag"`
}
