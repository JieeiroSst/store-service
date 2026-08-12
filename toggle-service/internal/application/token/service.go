package token

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	tokens port.TokenRepository
}

func NewService(tokens port.TokenRepository) port.TokenService {
	return &service{tokens: tokens}
}

func (s *service) Create(ctx context.Context, in port.CreateTokenInput, actor string) (string, *model.APIToken, error) {
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		return "", nil, err
	}

	projectPart := "*"
	if in.ProjectID != nil {
		projectPart = in.ProjectID.String()
	}
	envPart := "*"
	if in.EnvironmentID != nil {
		envPart = in.EnvironmentID.String()
	}
	plaintext := fmt.Sprintf("%s:%s.%s.%s", in.Type, projectPart, envPart, hex.EncodeToString(random))

	hash := sha256.Sum256([]byte(plaintext))
	hashHex := hex.EncodeToString(hash[:])

	prefix := plaintext
	if len(prefix) > 16 {
		prefix = prefix[:16]
	}

	t := &model.APIToken{
		Name:          in.Name,
		TokenHash:     hashHex,
		TokenPrefix:   prefix,
		Type:          in.Type,
		ProjectID:     in.ProjectID,
		EnvironmentID: in.EnvironmentID,
		ExpiresAt:     in.ExpiresAt,
		CreatedBy:     actor,
	}
	if err := s.tokens.Create(ctx, t); err != nil {
		return "", nil, err
	}
	return plaintext, t, nil
}

func (s *service) Resolve(ctx context.Context, plaintext string) (*model.APIToken, error) {
	hash := sha256.Sum256([]byte(plaintext))
	hashHex := hex.EncodeToString(hash[:])

	t, err := s.tokens.GetByHash(ctx, hashHex)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, apperr.ErrUnauthorized
	}
	if t.ExpiresAt != nil && t.ExpiresAt.Before(time.Now()) {
		return nil, apperr.ErrUnauthorized
	}
	return t, nil
}

func (s *service) List(ctx context.Context) ([]model.APIToken, error) {
	return s.tokens.List(ctx)
}

func (s *service) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.tokens.Delete(ctx, id)
}
