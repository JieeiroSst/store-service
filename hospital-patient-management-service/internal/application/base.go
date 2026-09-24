package application

import (
	"context"
	"math"
	"net/mail"
	"strings"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type base[T any] struct {
	repo port.Store[T]
}

func (b base[T]) Get(ctx context.Context, id int32) (*T, error) {
	if id <= 0 {
		return nil, model.Invalid("id is required")
	}
	return b.repo.Get(ctx, id)
}

func (b base[T]) Delete(ctx context.Context, id int32) error {
	if id <= 0 {
		return model.Invalid("id is required")
	}
	return b.repo.Delete(ctx, id)
}

func update[T any](ctx context.Context, repo port.Store[T], id int32, v *T) (*T, error) {
	if id <= 0 {
		return nil, model.Invalid("id is required")
	}
	if err := repo.Update(ctx, v); err != nil {
		return nil, err
	}
	return repo.Get(ctx, id)
}

func create[T any](ctx context.Context, repo port.Store[T], v *T) (*T, error) {
	if err := repo.Create(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func required(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return model.Invalid("%s is required", field)
	}
	return nil
}

func positive(field string, id int32) error {
	if id <= 0 {
		return model.Invalid("%s is required", field)
	}
	return nil
}

func validEmail(field, value string) error {
	if _, err := mail.ParseAddress(value); err != nil {
		return model.Invalid("%s is not a valid email address", field)
	}
	return nil
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func money(v float64) float64 { return math.Round(v*100) / 100 }

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
