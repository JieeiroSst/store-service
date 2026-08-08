package model

type IncomingMessage struct {
	Channel      Channel
	ChatID       string
	MessageID    string
	FromUsername string
	Text         string
}

type OutgoingMessage struct {
	Channel          Channel
	ChatID           string
	Text             string
	ReplyToMessageID string
}
