package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"github.com/JIeeiroSst/customer-info-service/internal/port"
)

type CustomerService struct {
	customers port.CustomerRepository
	locker    port.Locker
	users     port.UserDirectory
	ekyc      port.EkycGateway
	hasher    port.DocumentHasher
	clock     port.Clock
	metrics   port.Metrics
	policy    domain.KYCPolicy
	random    io.Reader
}

func NewCustomerService(policy domain.KYCPolicy, customers port.CustomerRepository, locker port.Locker, users port.UserDirectory, ekyc port.EkycGateway, hasher port.DocumentHasher, clock port.Clock, metrics port.Metrics) *CustomerService {
	return &CustomerService{
		customers: customers,
		locker:    locker,
		users:     users,
		ekyc:      ekyc,
		hasher:    hasher,
		clock:     clock,
		metrics:   metrics,
		policy:    policy,
		random:    rand.Reader,
	}
}

func (s *CustomerService) Onboard(ctx context.Context, userID int64) (*domain.Customer, bool, error) {
	if userID <= 0 {
		return nil, false, domain.Invalid("user_id must be a positive integer")
	}
	var (
		out     *domain.Customer
		created bool
	)
	err := s.locker.WithLock(ctx, userLockKey(userID), func(ctx context.Context) error {
		existing, err := s.customers.GetByUserID(ctx, userID)
		if err == nil {
			out = existing
			return nil
		}
		if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		profile, err := s.users.GetUser(ctx, userID)
		if err != nil {
			return err
		}
		id, err := newID(s.random)
		if err != nil {
			return err
		}
		c, err := domain.NewCustomer(id, profile, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.customers.Create(ctx, c); err != nil {
			return err
		}
		out, created = c, true
		return nil
	})
	if err != nil {
		s.metrics.ObserveOnboarding("failed")
		return nil, false, err
	}
	if created {
		s.metrics.ObserveOnboarding("created")
	} else {
		s.metrics.ObserveOnboarding("existing")
	}
	return out, created, nil
}

func (s *CustomerService) Get(ctx context.Context, id string) (*domain.Customer, error) {
	return s.customers.Get(ctx, id)
}

func (s *CustomerService) GetByUserID(ctx context.Context, userID int64) (*domain.Customer, error) {
	if userID <= 0 {
		return nil, domain.Invalid("user_id must be a positive integer")
	}
	return s.customers.GetByUserID(ctx, userID)
}

func (s *CustomerService) Sync(ctx context.Context, id string) (*domain.Customer, error) {
	var out *domain.Customer
	err := s.locker.WithLock(ctx, customerLockKey(id), func(ctx context.Context) error {
		c, err := s.customers.Get(ctx, id)
		if err != nil {
			return err
		}
		profile, err := s.users.GetUser(ctx, c.UserID)
		if err != nil {
			return err
		}
		c.Sync(profile, s.clock.Now())
		out = c
		return s.customers.Update(ctx, c)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *CustomerService) SubmitKYC(ctx context.Context, id string, front, back []byte) (*domain.Customer, error) {
	for name, img := range map[string][]byte{"front": front, "back": back} {
		if len(img) == 0 || len(img) > domain.MaxDocumentImage {
			return nil, domain.Invalid("%s image must be between 1 byte and %d MB", name, domain.MaxDocumentImage>>20)
		}
	}
	c, err := s.activeCustomer(ctx, id)
	if err != nil {
		return nil, err
	}
	res, err := s.ekyc.SubmitCitizenCard(ctx, c.UserID, front, back)
	if err != nil {
		return nil, err
	}
	return s.applyKYC(ctx, id, res)
}

func (s *CustomerService) RefreshKYC(ctx context.Context, id string) (*domain.Customer, error) {
	c, err := s.activeCustomer(ctx, id)
	if err != nil {
		return nil, err
	}
	res, err := s.ekyc.Status(ctx, c.UserID)
	if err != nil {
		return nil, err
	}
	return s.applyKYC(ctx, id, res)
}

func (s *CustomerService) activeCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	c, err := s.customers.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != domain.CustomerActive {
		return nil, domain.Conflict("customer is %s", c.Status)
	}
	return c, nil
}

func (s *CustomerService) applyKYC(ctx context.Context, id string, res domain.EkycResult) (*domain.Customer, error) {
	var out *domain.Customer
	err := s.locker.WithLock(ctx, customerLockKey(id), func(ctx context.Context) error {
		c, err := s.customers.Get(ctx, id)
		if err != nil {
			return err
		}
		if c.Status != domain.CustomerActive {
			return domain.Conflict("customer is %s", c.Status)
		}
		kyc := s.policy.Evaluate(c, res, s.clock.Now())
		if kyc.Document.Number != "" {
			kyc.DocumentHash = s.hasher.Hash(kyc.Document.Number)
		}
		if kyc.Status == domain.KYCVerified {
			if err := s.ensureDocumentUnused(ctx, kyc.DocumentHash, c.ID); err != nil {
				return err
			}
		}
		c.KYC = kyc
		c.UpdatedAt = s.clock.Now()
		out = c
		return s.customers.Update(ctx, c)
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveKYC(out.KYC.Status)
	return out, nil
}

func (s *CustomerService) ensureDocumentUnused(ctx context.Context, hash, customerID string) error {
	other, err := s.customers.GetByDocumentHash(ctx, hash)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return nil
	case err != nil:
		return err
	case other.ID != customerID:
		return domain.Conflict("this identity document is already verified for another customer")
	}
	return nil
}

func userLockKey(id int64) string { return "user:" + strconv.FormatInt(id, 10) }

func customerLockKey(id string) string { return "customer:" + id }

func newID(r io.Reader) (string, error) {
	var b [16]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
