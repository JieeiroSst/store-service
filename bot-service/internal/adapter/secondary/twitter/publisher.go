package twitter

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
)

type contentPublisher struct {
	client *Client
}

func NewContentPublisher(client *Client) port.ChannelPublisher {
	return &contentPublisher{client: client}
}

func (p *contentPublisher) Channel() model.Channel { return model.ChannelTwitter }

func (p *contentPublisher) Publish(ctx context.Context, post model.Post) (string, error) {
	var mediaID string
	if len(post.Media) > 0 {
		id, err := p.client.UploadMedia(ctx, post.Media[0].URL)
		if err != nil {
			return "", fmt.Errorf("twitter: upload media: %w", err)
		}
		mediaID = id
	}
	return p.client.CreateTweet(ctx, post.Text, mediaID)
}
