package domain

import "strings"

type IntentionAction string

const (
	IntentionAllow IntentionAction = "allow"
	IntentionDeny  IntentionAction = "deny"
)

const Wildcard = "*"

type Intention struct {
	ID              string
	SourceName      string
	DestinationName string
	Action          IntentionAction
	Description     string
	Meta            map[string]string
	Precedence      int
	RaftIndex
}

func (i *Intention) Validate() error {
	i.SourceName = strings.TrimSpace(i.SourceName)
	i.DestinationName = strings.TrimSpace(i.DestinationName)
	if i.SourceName == "" || i.DestinationName == "" {
		return Invalid("SourceName and DestinationName are required")
	}
	if i.Action != IntentionAllow && i.Action != IntentionDeny {
		return Invalid("Action must be %q or %q", IntentionAllow, IntentionDeny)
	}
	i.Precedence = IntentionPrecedence(i.SourceName, i.DestinationName)
	return nil
}

func IntentionPrecedence(src, dst string) int {
	switch {
	case dst != Wildcard && src != Wildcard:
		return 9
	case dst != Wildcard:
		return 8
	case src != Wildcard:
		return 6
	default:
		return 5
	}
}

func (i Intention) Matches(src, dst string) bool {
	return (i.SourceName == Wildcard || i.SourceName == src) &&
		(i.DestinationName == Wildcard || i.DestinationName == dst)
}
