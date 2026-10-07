package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
)

type AccountRepository interface {
	Create(ctx context.Context, a *domain.Account) error
	Update(ctx context.Context, a *domain.Account) error
	Get(ctx context.Context, id string) (*domain.Account, error)
	ListByCustomer(ctx context.Context, customerID string) ([]*domain.Account, error)
}

type CardRepository interface {
	Create(ctx context.Context, c *domain.Card) error
	Update(ctx context.Context, c *domain.Card) error
	Get(ctx context.Context, id string) (*domain.Card, error)
	GetByPANHash(ctx context.Context, hash string) (*domain.Card, error)
	ListByAccount(ctx context.Context, accountID string) ([]*domain.Card, error)
}

type AuthorizationRepository interface {
	Create(ctx context.Context, a *domain.Authorization) error
	Update(ctx context.Context, a *domain.Authorization) error
	Get(ctx context.Context, id string) (*domain.Authorization, error)
	ListByCard(ctx context.Context, cardID string, limit int) ([]*domain.Authorization, error)
	SumDebits(ctx context.Context, cardID string, since time.Time) (int64, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, t *domain.Transaction) error
	ListByAccount(ctx context.Context, accountID string, limit int) ([]*domain.Transaction, error)
}

type Locker interface {
	WithLock(ctx context.Context, key string, fn func(ctx context.Context) error) error
}

type CustomerDirectory interface {
	GetCustomer(ctx context.Context, id string) (domain.Customer, error)
}

type PaymentAuthenticator interface {
	Enabled() bool
	Consume(ctx context.Context, authenticationID, cardID string, amount int64, currency string) error
}

type Vault interface {
	HashPAN(pan string) string
	EncryptPAN(pan string) ([]byte, error)
	DecryptPAN(cipher []byte) (string, error)
}

type CardSecurity interface {
	CVV(pan string, expiry domain.Expiry, serviceCode string) string
	HashPIN(pin string) (string, error)
	VerifyPIN(pin, hash string) bool
}

type Clock interface {
	Now() time.Time
}

type Metrics interface {
	ObserveIssued(program string, cardType domain.CardType)
	ObserveLifecycle(event string)
	ObserveAuthorization(channel string, code domain.ResponseCode)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
