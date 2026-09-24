package usecase

import (
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/car-rental-service/internal/repository"
	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
)

const (
	defaultPageSize = 20
	maxPageSize     = 100
)

type Usecase struct {
	repos   *repository.Repositories
	gateway PaymentGateway
	pricing PricingPolicy
}

type Dependency struct {
	Repos   *repository.Repositories
	Gateway PaymentGateway
	Pricing PricingPolicy
}

func NewUsecase(deps Dependency) *Usecase {
	if deps.Gateway == nil {
		deps.Gateway = ManualGateway{}
	}
	return &Usecase{repos: deps.Repos, gateway: deps.Gateway, pricing: deps.Pricing.withDefaults()}
}

type Page struct {
	Offset, Limit int
}

func NewPage(size int32, token string) (Page, error) {
	p := Page{Limit: int(size)}
	if p.Limit <= 0 {
		p.Limit = defaultPageSize
	}
	if p.Limit > maxPageSize {
		p.Limit = maxPageSize
	}
	if token != "" {
		off, err := strconv.Atoi(token)
		if err != nil || off < 0 {
			return p, fmt.Errorf("%w: invalid page token", model.ErrInvalidArgument)
		}
		p.Offset = off
	}
	return p, nil
}

func (p Page) NextToken(n int, total int64) string {
	next := p.Offset + n
	if n == 0 || int64(next) >= total {
		return ""
	}
	return strconv.Itoa(next)
}

func parseID(field, s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a valid UUID", model.ErrInvalidArgument, field)
	}
	return id, nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", model.ErrInvalidArgument, fmt.Sprintf(format, args...))
}

func precondition(format string, args ...any) error {
	return fmt.Errorf("%w: %s", model.ErrFailedPrecond, fmt.Sprintf(format, args...))
}
