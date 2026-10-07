package application

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"io"
	"log"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/JIeeiroSst/card-service/internal/port"
)

type Settings struct {
	LimitsLocation *time.Location
}

type AuthorizationService struct {
	catalog       *domain.Catalog
	accounts      port.AccountRepository
	cards         port.CardRepository
	auths         port.AuthorizationRepository
	transactions  port.TransactionRepository
	authenticator port.PaymentAuthenticator
	locker        port.Locker
	vault         port.Vault
	security      port.CardSecurity
	clock         port.Clock
	metrics       port.Metrics
	location      *time.Location
	random        io.Reader
}

func NewAuthorizationService(settings Settings, catalog *domain.Catalog, accounts port.AccountRepository, cards port.CardRepository, auths port.AuthorizationRepository, transactions port.TransactionRepository, authenticator port.PaymentAuthenticator, locker port.Locker, vault port.Vault, security port.CardSecurity, clock port.Clock, metrics port.Metrics) *AuthorizationService {
	loc := settings.LimitsLocation
	if loc == nil {
		loc = time.UTC
	}
	return &AuthorizationService{
		catalog: catalog, accounts: accounts, cards: cards, auths: auths, transactions: transactions,
		authenticator: authenticator, locker: locker, vault: vault, security: security, clock: clock,
		metrics: metrics, location: loc, random: rand.Reader,
	}
}

func (s *AuthorizationService) Authorize(ctx context.Context, req domain.AuthorizationRequest) (*domain.Authorization, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	found, err := s.cards.GetByPANHash(ctx, s.vault.HashPAN(req.PAN))
	if errors.Is(err, domain.ErrNotFound) {
		s.metrics.ObserveAuthorization(string(req.Channel), domain.CodeInvalidCard)
		now := s.clock.Now()
		return &domain.Authorization{
			ProcessingCode: req.ProcessingCode, Status: domain.AuthDenied, Amount: req.Amount, Currency: req.Currency,
			Channel: req.Channel, Merchant: req.Merchant, MCC: req.MCC, Code: domain.CodeInvalidCard,
			Reason: "unknown card", CreatedAt: now, UpdatedAt: now,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	program, err := s.catalog.Get(found.ProgramCode)
	if err != nil {
		return nil, err
	}
	stepUp := s.stepUp(ctx, found, program, req)

	var auth *domain.Authorization
	err = s.locker.WithLock(ctx, accountLockKey(found.AccountID), func(ctx context.Context) error {
		return s.locker.WithLock(ctx, cardLockKey(found.ID), func(ctx context.Context) error {
			card, err := s.cards.Get(ctx, found.ID)
			if err != nil {
				return err
			}
			account, err := s.accounts.Get(ctx, card.AccountID)
			if err != nil {
				return err
			}
			now := s.clock.Now()
			spent, err := s.auths.SumDebits(ctx, card.ID, s.dayStart(now))
			if err != nil {
				return err
			}
			cred := s.credentials(card, req)
			cred.StepUp = stepUp
			decision := domain.Decide(card, account, program, req, cred, spent, now)
			if err := s.trackPIN(ctx, card, cred, decision, now); err != nil {
				return err
			}
			id, err := newID(s.random)
			if err != nil {
				return err
			}
			auth = domain.NewAuthorization(id, card, req, decision, now)
			if decision.Approved() {
				if auth.AuthCode, err = domain.NewAuthCode(s.random); err != nil {
					return err
				}
				if err := s.applyApproval(ctx, account, auth, now); err != nil {
					return err
				}
			}
			return s.auths.Create(ctx, auth)
		})
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveAuthorization(string(auth.Channel), auth.Code)
	return auth, nil
}

func (s *AuthorizationService) stepUp(ctx context.Context, card *domain.Card, p domain.Program, req domain.AuthorizationRequest) domain.Check {
	if !req.NeedsStepUp(p) || !s.authenticator.Enabled() {
		return domain.CheckMatch
	}
	if req.AuthenticationID == "" {
		return domain.CheckNotProvided
	}
	if err := s.authenticator.Consume(ctx, req.AuthenticationID, card.ID, req.Amount, req.Currency); err != nil {
		if !domain.IsConflict(err) && !errors.Is(err, domain.ErrNotFound) {
			log.Printf("payment authentication %s could not be checked: %v", req.AuthenticationID, err)
		}
		return domain.CheckMismatch
	}
	return domain.CheckMatch
}

func (s *AuthorizationService) trackPIN(ctx context.Context, card *domain.Card, cred domain.Credentials, d domain.Decision, now time.Time) error {
	switch {
	case d.Code == domain.CodeIncorrectPIN && cred.PIN == domain.CheckMismatch:
		card.RecordPINFailure(now)
	case cred.PIN == domain.CheckMatch && card.PINFailures > 0 && !card.PINLocked():
		card.RecordPINSuccess(now)
	default:
		return nil
	}
	return s.cards.Update(ctx, card)
}

func (s *AuthorizationService) applyApproval(ctx context.Context, account *domain.Account, auth *domain.Authorization, now time.Time) error {
	if auth.ProcessingCode.Debit() {
		account.Hold(auth.Amount, now)
		return s.accounts.Update(ctx, account)
	}
	account.Post(auth.Amount, now)
	if err := s.accounts.Update(ctx, account); err != nil {
		return err
	}
	return s.post(ctx, account, auth, auth.Amount, now)
}

func (s *AuthorizationService) Confirm(ctx context.Context, id string, amount int64) (*domain.Authorization, error) {
	return s.settle(ctx, id, "authorization_confirm", func(a *domain.Authorization, account *domain.Account, now time.Time) error {
		release, err := a.Confirm(amount, now)
		if err != nil {
			return err
		}
		account.Release(release, now)
		account.Post(-a.ConfirmedAmount, now)
		return s.post(ctx, account, a, -a.ConfirmedAmount, now)
	})
}

func (s *AuthorizationService) Cancel(ctx context.Context, id string) (*domain.Authorization, error) {
	return s.settle(ctx, id, "authorization_cancel", func(a *domain.Authorization, account *domain.Account, now time.Time) error {
		release, err := a.Cancel(now)
		if err != nil {
			return err
		}
		account.Release(release, now)
		return nil
	})
}

func (s *AuthorizationService) settle(ctx context.Context, id, event string, fn func(a *domain.Authorization, account *domain.Account, now time.Time) error) (*domain.Authorization, error) {
	current, err := s.auths.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	var out *domain.Authorization
	err = s.locker.WithLock(ctx, accountLockKey(current.AccountID), func(ctx context.Context) error {
		a, err := s.auths.Get(ctx, id)
		if err != nil {
			return err
		}
		account, err := s.accounts.Get(ctx, a.AccountID)
		if err != nil {
			return err
		}
		if err := fn(a, account, s.clock.Now()); err != nil {
			return err
		}
		if err := s.accounts.Update(ctx, account); err != nil {
			return err
		}
		out = a
		return s.auths.Update(ctx, a)
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveLifecycle(event)
	return out, nil
}

func (s *AuthorizationService) post(ctx context.Context, account *domain.Account, a *domain.Authorization, amount int64, now time.Time) error {
	id, err := newID(s.random)
	if err != nil {
		return err
	}
	return s.transactions.Create(ctx, &domain.Transaction{
		ID: id, AccountID: account.ID, CardID: a.CardID, AuthorizationID: a.ID,
		Type: a.ProcessingCode.TransactionType(), ProcessingCode: a.ProcessingCode, Amount: amount,
		BalanceAfter: account.Balance, Description: a.Merchant, CreatedAt: now,
	})
}

func (s *AuthorizationService) Get(ctx context.Context, id string) (*domain.Authorization, error) {
	return s.auths.Get(ctx, id)
}

func (s *AuthorizationService) ListByCard(ctx context.Context, cardID string, limit int) ([]*domain.Authorization, error) {
	if _, err := s.cards.Get(ctx, cardID); err != nil {
		return nil, err
	}
	return s.auths.ListByCard(ctx, cardID, clampLimit(limit))
}

func (s *AuthorizationService) SpentToday(ctx context.Context, cardID string) (int64, error) {
	return s.auths.SumDebits(ctx, cardID, s.dayStart(s.clock.Now()))
}

func (s *AuthorizationService) credentials(card *domain.Card, req domain.AuthorizationRequest) domain.Credentials {
	var cred domain.Credentials
	if req.CVV != "" {
		want := s.security.CVV(req.PAN, card.Expiry, card.ServiceCode)
		cred.CVV = domain.CheckMismatch
		if subtle.ConstantTimeCompare([]byte(req.CVV), []byte(want)) == 1 {
			cred.CVV = domain.CheckMatch
		}
	}
	if req.PIN != "" {
		cred.PIN = domain.CheckMismatch
		if card.PINSet() && s.security.VerifyPIN(req.PIN, card.PINHash) {
			cred.PIN = domain.CheckMatch
		}
	}
	return cred
}

func (s *AuthorizationService) dayStart(now time.Time) time.Time {
	t := now.In(s.location)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, s.location)
}
