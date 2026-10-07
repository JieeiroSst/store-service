package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

type CustomerRepository interface {
	Create(ctx context.Context, c *domain.Customer) error
	Update(ctx context.Context, c *domain.Customer) error
	Get(ctx context.Context, id string) (*domain.Customer, error)
	GetByUserID(ctx context.Context, userID int64) (*domain.Customer, error)
	GetByDocumentHash(ctx context.Context, hash string) (*domain.Customer, error)
}

type Locker interface {
	WithLock(ctx context.Context, key string, fn func(ctx context.Context) error) error
}

type UserDirectory interface {
	GetUser(ctx context.Context, id int64) (domain.UserProfile, error)
}

type EkycGateway interface {
	SubmitCitizenCard(ctx context.Context, userID int64, front, back []byte) (domain.EkycResult, error)
	Status(ctx context.Context, userID int64) (domain.EkycResult, error)
}

type DocumentHasher interface {
	Hash(documentNumber string) string
}

type Clock interface {
	Now() time.Time
}

type Metrics interface {
	ObserveOnboarding(result string)
	ObserveKYC(status domain.KYCStatus)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
