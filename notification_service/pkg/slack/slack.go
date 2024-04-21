package slack

import (
	"fmt"

	"github.com/ashwanthkumar/slack-go-webhook"
)

type slackHook struct {
	serect string
}

type Slack interface {
	PushNoti(payload PayloadSlack) error
}

type PayloadSlack struct {
	Channel    string
	Username   string
	Text       string
	IconEmoji  string
	Attachment Attachment
}

type Attachment struct {
	Type     string
	Text     string
	DeepLink string
}

func NewSlack(serect string) Slack {
	return &slackHook{serect: serect}
}

func (p *slackHook) PushNoti(payload PayloadSlack) error {
	webhookUrl := fmt.Sprintf("https://hooks.slack.com/services/%v", p.serect)

	attachment1 := slack.Attachment{}
	attachment1.AddAction(slack.Action{Type: payload.Attachment.Type, Text: payload.Attachment.Text, Url: payload.Attachment.DeepLink, Style: "primary"})
	data := slack.Payload{
		Text:        payload.Text,
		Username:    payload.Username,
		Channel:     payload.Channel,
		IconEmoji:   payload.IconEmoji,
		Attachments: []slack.Attachment{attachment1},
	}
	err := slack.Send(webhookUrl, "", data)
	if len(err) > 0 {
		fmt.Printf("error: %s\n", err)
		return err[0]
	}
	return nil
}
