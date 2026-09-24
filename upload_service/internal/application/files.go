package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/JIeeiroSst/upload-service/config"
	"github.com/JIeeiroSst/upload-service/internal/domain/model"
	"github.com/JIeeiroSst/upload-service/internal/domain/port"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const (
	maxNameLen = 255
	maxLimit   = 100
	defLimit   = 20
)

type FileService struct {
	meta    port.MetadataRepository
	objects port.ObjectStore
	clock   port.Clock
	max     int64
	allowed []string
}

func NewFileService(meta port.MetadataRepository, objects port.ObjectStore, clock port.Clock, cfg *config.Config) port.FileUsecase {
	return &FileService{meta: meta, objects: objects, clock: clock, max: cfg.Upload.MaxBytes, allowed: cfg.Upload.AllowedTypes}
}

type prepared struct {
	name, contentType, sum string
	data                   []byte
}

func (s *FileService) prepare(in port.Upload) (*prepared, error) {
	data, err := io.ReadAll(io.LimitReader(in.Body, s.max+1))
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	if int64(len(data)) > s.max {
		return nil, fmt.Errorf("%w: limit is %d MB", model.ErrTooLarge, s.max>>20)
	}
	if len(data) == 0 {
		return nil, model.Invalid("file is empty")
	}
	ct := http.DetectContentType(data)
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	if !slices.Contains(s.allowed, ct) {
		return nil, fmt.Errorf("%w: %s", model.ErrUnsupportedType, ct)
	}
	sum := sha256.Sum256(data)
	return &prepared{name: cleanName(in.FileName), contentType: ct, sum: hex.EncodeToString(sum[:]), data: data}, nil
}

func cleanName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, `\`, "/"))
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
	if name == "." || name == ".." || name == "/" || strings.TrimSpace(name) == "" {
		return "file"
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	return name
}

func (s *FileService) put(ctx context.Context, key string, p *prepared) error {
	return s.objects.Put(ctx, key, bytes.NewReader(p.data), int64(len(p.data)), p.contentType)
}

func (s *FileService) Create(ctx context.Context, in port.Upload) (*model.File, error) {
	in.ReceiverID = strings.TrimSpace(in.ReceiverID)
	if in.ReceiverID == "" {
		return nil, model.Invalid("receiver_id is required")
	}
	p, err := s.prepare(in)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()
	f := &model.File{
		ID: uuid.NewString(), ReceiverID: in.ReceiverID, FileName: p.name, ContentType: p.contentType,
		Size: int64(len(p.data)), SHA256: p.sum, CreatedAt: now, UpdatedAt: now, ObjectKey: uuid.NewString(),
	}
	if err := s.put(ctx, f.ObjectKey, p); err != nil {
		return nil, err
	}
	if err := s.meta.Insert(ctx, f); err != nil {
		// Do not leave an unreachable copy of the file behind.
		s.discard(ctx, f.ObjectKey)
		return nil, err
	}
	return f, nil
}

func (s *FileService) Replace(ctx context.Context, id string, in port.Upload) (*model.File, error) {
	f, err := s.meta.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	p, err := s.prepare(in)
	if err != nil {
		return nil, err
	}

	oldKey := f.ObjectKey
	f.ObjectKey, f.FileName, f.ContentType = uuid.NewString(), p.name, p.contentType
	f.Size, f.SHA256, f.UpdatedAt = int64(len(p.data)), p.sum, s.clock.Now()
	if err := s.put(ctx, f.ObjectKey, p); err != nil {
		return nil, err
	}
	if err := s.meta.ReplaceContent(ctx, f); err != nil {
		s.discard(ctx, f.ObjectKey)
		return nil, err
	}
	s.discard(ctx, oldKey)
	return f, nil
}

func (s *FileService) Get(ctx context.Context, id string) (*model.File, error) {
	return s.meta.Get(ctx, id)
}

func (s *FileService) Open(ctx context.Context, id string) (*port.Download, error) {
	f, err := s.meta.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	body, err := s.objects.Get(ctx, f.ObjectKey)
	if err != nil {
		return nil, err
	}
	return &port.Download{File: f, Body: body}, nil
}

func (s *FileService) List(ctx context.Context, receiverID string, limit, offset int) ([]model.File, int64, error) {
	if strings.TrimSpace(receiverID) == "" {
		return nil, 0, model.Invalid("receiver_id is required")
	}
	if limit <= 0 {
		limit = defLimit
	}
	limit = min(limit, maxLimit)
	return s.meta.List(ctx, receiverID, limit, max(offset, 0))
}

func (s *FileService) Delete(ctx context.Context, id string) error {
	f, err := s.meta.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.objects.Delete(ctx, f.ObjectKey); err != nil {
		return err
	}
	return s.meta.Delete(ctx, id)
}

func (s *FileService) discard(ctx context.Context, key string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()
	if err := s.objects.Delete(ctx, key); err != nil && !errors.Is(err, model.ErrNotFound) {
		logrus.WithError(err).WithField("object_key", key).Error("could not remove an orphaned object")
	}
}
