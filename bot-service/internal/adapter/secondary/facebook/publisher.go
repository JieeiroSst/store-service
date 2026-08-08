package facebook

import (
	"context"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
)

type contentPublisher struct {
	client *Client
}

func NewContentPublisher(client *Client) port.ChannelPublisher {
	return &contentPublisher{client: client}
}

func (p *contentPublisher) Channel() model.Channel { return model.ChannelFacebook }

func (p *contentPublisher) Publish(ctx context.Context, post model.Post) (string, error) {
	var imageURL string
	if len(post.Media) > 0 {
		imageURL = post.Media[0].URL
	}
	return p.client.PostToFeed(ctx, post.Text, imageURL)
}
