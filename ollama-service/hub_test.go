package main

import (
	"testing"
	"time"
)

func newTestClient(id int64) *Client {
	return &Client{SendChan: make(chan Message, 4), UserID: id}
}

func received(c *Client) []Message {
	var out []Message
	for {
		select {
		case m, ok := <-c.SendChan:
			if !ok {
				return out
			}
			out = append(out, m)
		default:
			return out
		}
	}
}

func TestRouteDeliversOnlyToParticipants(t *testing.T) {
	prevMembers, prevDB := groupMembersOf, saveMessage
	groupMembersOf = func(int64) (map[int64]bool, error) { return map[int64]bool{1: true, 3: true}, nil }
	saveMessage = func(Message) {}
	t.Cleanup(func() { groupMembersOf, saveMessage = prevMembers, prevDB })

	server := NewChatServer()
	a, b, c := newTestClient(1), newTestClient(2), newTestClient(3)
	for _, cl := range []*Client{a, b, c} {
		server.Clients[cl] = true
	}

	server.route(Message{SenderID: "1", RecipientID: "2", Timestamp: time.Now()})
	if len(received(a)) != 1 || len(received(b)) != 1 || len(received(c)) != 0 {
		t.Fatal("direct message must reach only sender and recipient")
	}

	server.route(Message{SenderID: "1", GroupID: 9})
	if len(received(a)) != 1 || len(received(b)) != 0 || len(received(c)) != 1 {
		t.Fatal("group message must reach only current members")
	}

	server.route(Message{SenderID: "1"})
	if len(received(a))+len(received(b))+len(received(c)) != 0 {
		t.Fatal("message without a target must not be broadcast")
	}
}

func TestSlowClientIsDroppedOnce(t *testing.T) {
	prev := saveMessage
	saveMessage = func(Message) {}
	t.Cleanup(func() { saveMessage = prev })

	server := NewChatServer()
	slow := &Client{SendChan: make(chan Message), UserID: 2}
	server.Clients[slow] = true

	server.route(Message{SenderID: "1", RecipientID: "2"})
	server.drop(slow)
	if _, ok := server.Clients[slow]; ok {
		t.Fatal("slow client should be removed")
	}
	if _, ok := <-slow.SendChan; ok {
		t.Fatal("send channel should be closed")
	}
}
