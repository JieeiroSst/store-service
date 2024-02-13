package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"strings"

	"github.com/google/uuid"
)

func NewAuthorizeGenerate() *AuthorizeTokenGenerate {
	return &AuthorizeTokenGenerate{}
}

// AuthorizeGenerate generate the authorize code
type AuthorizeTokenGenerate struct{}

// Token based on the UUID generated token
func (ag *AuthorizeTokenGenerate) Token(ctx context.Context, data *GenerateBasic) (string, error) {
	buf := bytes.NewBufferString(data.Client.GetID())
	buf.WriteString(data.UserID)
	token := uuid.NewMD5(uuid.Must(uuid.NewRandom()), buf.Bytes())
	code := base64.URLEncoding.EncodeToString([]byte(token.String()))
	code = strings.ToUpper(strings.TrimRight(code, "="))

	return code, nil
}
