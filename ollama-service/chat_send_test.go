package main

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func withMembership(t *testing.T, member bool) {
	t.Helper()
	prev := groupMembershipFor
	groupMembershipFor = func(int64, int64) bool { return member }
	t.Cleanup(func() { groupMembershipFor = prev })
}

func TestBuildOutgoingMessageDirect(t *testing.T) {
	now := time.Unix(1700000000, 0)
	msg, err := buildOutgoingMessage(7, SendMessageRequest{RecipientID: 9, Content: "  chào  "}, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "user" || msg.RecipientID != "9" || msg.SenderID != "7" || msg.GroupID != 0 {
		t.Fatalf("unexpected routing: %+v", msg)
	}
	if msg.MessageType != "text" || msg.Data.Content != "chào" || !msg.Timestamp.Equal(now) {
		t.Fatalf("unexpected payload: %+v", msg)
	}
}

func TestBuildOutgoingMessageGroup(t *testing.T) {
	withMembership(t, true)
	msg, err := buildOutgoingMessage(7, SendMessageRequest{GroupID: 3, MessageType: "image", URL: "https://cdn/x.png"}, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.Type != "group" || msg.GroupID != 3 || msg.RecipientID != "" || msg.Data.URL != "https://cdn/x.png" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestBuildOutgoingMessageRejects(t *testing.T) {
	cases := []struct {
		name   string
		member bool
		req    SendMessageRequest
		want   error
	}{
		{"no target", true, SendMessageRequest{Content: "x"}, errMessageTarget},
		{"both targets", true, SendMessageRequest{RecipientID: 2, GroupID: 3, Content: "x"}, errMessageTarget},
		{"empty", true, SendMessageRequest{RecipientID: 2, Content: "   "}, errMessageEmpty},
		{"too long", true, SendMessageRequest{RecipientID: 2, Content: strings.Repeat("a", maxMessageContent+1)}, errMessageTooLong},
		{"bad kind", true, SendMessageRequest{RecipientID: 2, Content: "x", MessageType: "exe"}, errMessageKind},
		{"self", true, SendMessageRequest{RecipientID: 7, Content: "x"}, errMessageSelf},
		{"not member", false, SendMessageRequest{GroupID: 3, Content: "x"}, errNotGroupMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withMembership(t, tc.member)
			if _, err := buildOutgoingMessage(7, tc.req, time.Now()); !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
		})
	}
}

func TestProcessMessageDropsGroupMessageFromNonMember(t *testing.T) {
	withMembership(t, false)
	server := NewChatServer()
	msg := Message{Type: "group", GroupID: 3, SenderID: "7"}
	server.processMessage(msg)
	if len(server.Broadcast) != 0 {
		t.Fatal("non-member group message was broadcast")
	}

	withMembership(t, true)
	server.processMessage(msg)
	if len(server.Broadcast) != 1 {
		t.Fatal("member group message was not broadcast")
	}
}
