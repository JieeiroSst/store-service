package notification

import (
	"encoding/json"
	"time"
)

type SlackClient struct {
	WebHookUrl string
	UserName   string
	Channel    string
	TimeOut    time.Duration
}

type SimpleSlackRequest struct {
	Text      string
	IconEmoji string
}

type SlackJobNotification struct {
	Color     string
	IconEmoji string
	Details   string
	Text      string
}

type SlackMessage struct {
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color         string      `json:"color,omitempty"`
	Fallback      string      `json:"fallback,omitempty"`
	CallbackID    string      `json:"callback_id,omitempty"`
	ID            int         `json:"id,omitempty"`
	AuthorID      string      `json:"author_id,omitempty"`
	AuthorName    string      `json:"author_name,omitempty"`
	AuthorSubname string      `json:"author_subname,omitempty"`
	AuthorLink    string      `json:"author_link,omitempty"`
	AuthorIcon    string      `json:"author_icon,omitempty"`
	Title         string      `json:"title,omitempty"`
	TitleLink     string      `json:"title_link,omitempty"`
	Pretext       string      `json:"pretext,omitempty"`
	Text          string      `json:"text,omitempty"`
	ImageURL      string      `json:"image_url,omitempty"`
	ThumbURL      string      `json:"thumb_url,omitempty"`
	MarkdownIn    []string    `json:"mrkdwn_in,omitempty"`
	Ts            json.Number `json:"ts,omitempty"`
}
