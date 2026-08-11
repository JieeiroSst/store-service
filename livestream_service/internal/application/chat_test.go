package application

import (
	"context"
	"testing"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
)

func TestChatPublishRejectsBannedUser(t *testing.T) {
	broadcaster := &fakeChatBroadcaster{}
	moderation := newFakeModerationStore()
	_ = moderation.Ban(context.Background(), "room-1", "troll", time.Hour)

	u := NewChatUsecase(broadcaster, moderation)

	err := u.Publish(context.Background(), model.ChatMessage{RoomID: "room-1", UserID: "troll", Body: "spam"})
	if err != ErrBanned {
		t.Fatalf("Publish() error = %v, want ErrBanned", err)
	}
	if len(broadcaster.published) != 0 {
		t.Error("expected the banned user's message to never reach the broadcaster")
	}
}

func TestChatPublishAllowsUnbannedUser(t *testing.T) {
	broadcaster := &fakeChatBroadcaster{}
	moderation := newFakeModerationStore()
	u := NewChatUsecase(broadcaster, moderation)

	msg := model.ChatMessage{RoomID: "room-1", UserID: "regular-user", Body: "hello"}
	if err := u.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if len(broadcaster.published) != 1 || broadcaster.published[0].Body != "hello" {
		t.Errorf("expected message to reach the broadcaster, got %+v", broadcaster.published)
	}
}
