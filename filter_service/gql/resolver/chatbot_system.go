package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ChatbotQuery() generated.ChatbotQueryResolver { return &chatbotQueryResolver{r} }

type chatbotQueryResolver struct{ *Resolver }

func (r *chatbotQueryResolver) ConversationHistory(ctx context.Context, obj *model.ChatbotQuery, conversationID int, limit *int, offset *int) (*model.ChatbotMessageList, error) {
	return r.Clients.ChatbotSystem.ConversationHistory(ctx, conversationID, limit, offset)
}

func (r *chatbotQueryResolver) Conversations(ctx context.Context, obj *model.ChatbotQuery) (*model.ChatbotConversationList, error) {
	return r.Clients.ChatbotSystem.Conversations(ctx)
}
