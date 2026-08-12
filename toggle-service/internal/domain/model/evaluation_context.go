package model

import "time"

type EvaluationContext struct {
	UserID        string            `json:"userId"`
	SessionID     string            `json:"sessionId"`
	RemoteAddress string            `json:"remoteAddress"`
	AppName       string            `json:"appName"`
	Properties    map[string]string `json:"properties"`
	CurrentTime   time.Time         `json:"currentTime"`
}

func (c EvaluationContext) FieldValue(field string) (string, bool) {
	switch field {
	case "userId":
		return c.UserID, c.UserID != ""
	case "sessionId":
		return c.SessionID, c.SessionID != ""
	case "remoteAddress":
		return c.RemoteAddress, c.RemoteAddress != ""
	case "appName":
		return c.AppName, c.AppName != ""
	default:
		v, ok := c.Properties[field]
		return v, ok
	}
}
