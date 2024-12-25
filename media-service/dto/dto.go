package dto

import (
	"mime/multipart"
	"time"
)

type Subscription struct {
	SubscriptionID int    `json:"subscription_id"`
	Name           string `json:"name"`
	SubscribedFrom int    `json:"subscribed_from"`
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

type SearchVideo struct {
	Videos []Video `json:"videos"`
	Total  int64   `json:"total"`
	Page   int     `json:"page"`
	Size   int     `json:"size"`
	Pages  int     `json:"pages"`
}

type SearchVideoRequest struct {
	Query string `json:"query"`
	Page  int    `json:"page"`
	Size  int    `json:"size"`
}

func (s SearchVideoRequest) Build() SearchVideoRequest {
	if s.Page == 0 {
		s.Page = 1
	}
	if s.Size == 0 {
		s.Size = 20
	}
	return SearchVideoRequest{
		Query: s.Query,
		Page:  s.Page,
		Size:  s.Size,
	}
}
