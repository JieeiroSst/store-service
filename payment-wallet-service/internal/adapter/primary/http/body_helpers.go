package http

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
)

func peekBody(c *gin.Context) map[string]any {
	raw, err := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	out := map[string]any{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	_ = dec.Decode(&out)
	return out
}

func bodyString(c *gin.Context, key string) string {
	v, ok := peekBody(c)[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func setBodyField(c *gin.Context, key string, value any) {
	body := peekBody(c)
	if body == nil {
		body = map[string]any{}
	}
	body[key] = value
	raw, _ := json.Marshal(body)
	c.Request.Body = io.NopCloser(bytes.NewReader(raw))
	c.Request.ContentLength = int64(len(raw))
}
