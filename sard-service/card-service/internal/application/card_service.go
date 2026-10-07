package application

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
	"github.com/JIeeiroSst/card-service/internal/port"
)

type CardService struct {
	catalog  *domain.Catalog
	accounts port.AccountRepository
	cards    port.CardRepository
	locker   port.Locker
	vault    port.Vault
	security port.CardSecurity
	clock    port.Clock
	metrics  port.Metrics
	random   io.Reader
}

func NewCardService(catalog *domain.Catalog, accounts port.AccountRepository, cards port.CardRepository, locker port.Locker, vault port.Vault, security port.CardSecurity, clock port.Clock, metrics port.Metrics) *CardService {
	return &CardService{
		catalog: catalog, accounts: accounts, cards: cards, locker: locker, vault: vault,
		security: security, clock: clock, metrics: metrics, random: rand.Reader,
	}
}

type IssueCommand struct {
	AccountID   string
	Type        domain.CardType
	PrintedName string
}

type IssuedCard struct {
	Card *domain.Card
	PAN  string
	CVV  string
}

type ReportCommand struct {
	Reason  domain.Status
	Note    string
	Reissue bool
}

type ResolvedCard struct {
	Card    *domain.Card
	Account *domain.Account
}

func (s *CardService) Issue(ctx context.Context, cmd IssueCommand) (*IssuedCard, error) {
	var issued *IssuedCard
	err := s.locker.WithLock(ctx, accountLockKey(cmd.AccountID), func(ctx context.Context) error {
		account, err := s.accounts.Get(ctx, cmd.AccountID)
		if err != nil {
			return err
		}
		if account.Status != domain.AccountNormal {
			return domain.Conflict("account is %s", account.Status)
		}
		program, err := s.catalog.Get(account.ProgramCode)
		if err != nil {
			return err
		}
		if !program.Allows(cmd.Type) {
			return domain.Invalid("program %s does not issue %s cards", program.Code, cmd.Type)
		}
		if err := s.checkCardQuota(ctx, account, program, cmd.Type); err != nil {
			return err
		}
		name := cmd.PrintedName
		if name == "" {
			name = account.HolderName
		}
		name, err = domain.NormalizeCardholderName(name)
		if err != nil {
			return err
		}
		issued, err = s.issue(ctx, program, account, cmd.Type, name, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveIssued(issued.Card.ProgramCode, issued.Card.Type)
	return issued, nil
}

func (s *CardService) checkCardQuota(ctx context.Context, account *domain.Account, p domain.Program, t domain.CardType) error {
	cards, err := s.cards.ListByAccount(ctx, account.ID)
	if err != nil {
		return err
	}
	max := map[domain.CardType]int{domain.CardPlastic: 1, domain.CardVirtual: p.MaxActiveVirtual, domain.CardTemporary: 1}[t]
	n := 0
	now := s.clock.Now()
	for _, c := range cards {
		if c.Type == t && !c.Status.Terminal() && !c.Expired(now) {
			n++
		}
	}
	if n >= max {
		return domain.Conflict("account already has %d active %s card(s), the maximum allowed", n, t)
	}
	return nil
}

func (s *CardService) Get(ctx context.Context, id string) (*domain.Card, error) {
	return s.cards.Get(ctx, id)
}

func (s *CardService) ListByAccount(ctx context.Context, accountID string) ([]*domain.Card, error) {
	if _, err := s.accounts.Get(ctx, accountID); err != nil {
		return nil, err
	}
	return s.cards.ListByAccount(ctx, accountID)
}

func (s *CardService) Resolve(ctx context.Context, pan string, expiry domain.Expiry) (*ResolvedCard, error) {
	pan = domain.NormalizePAN(pan)
	if !domain.ValidPAN(pan) {
		return nil, domain.Invalid("pan must be %d digits with a valid Luhn check digit", domain.PANLength)
	}
	card, err := s.cards.GetByPANHash(ctx, s.vault.HashPAN(pan))
	if err != nil {
		return nil, err
	}
	if card.Expiry != expiry {
		return nil, domain.ErrNotFound
	}
	account, err := s.accounts.Get(ctx, card.AccountID)
	if err != nil {
		return nil, err
	}
	return &ResolvedCard{Card: card, Account: account}, nil
}

func (s *CardService) Activate(ctx context.Context, id, cvv string) (*domain.Card, error) {
	return s.mutate(ctx, id, "card_activate", func(c *domain.Card, now time.Time) error {
		if c.Status == domain.StatusPending {
			ok, err := s.cvvMatches(c, cvv)
			if err != nil {
				return err
			}
			if !ok {
				return domain.ErrCVVMismatch
			}
		}
		return c.Activate(now)
	})
}

func (s *CardService) Block(ctx context.Context, id, reason string) (*domain.Card, error) {
	return s.mutate(ctx, id, "card_block", func(c *domain.Card, now time.Time) error { return c.Block(reason, now) })
}

func (s *CardService) Unblock(ctx context.Context, id string) (*domain.Card, error) {
	return s.mutate(ctx, id, "card_unblock", func(c *domain.Card, now time.Time) error { return c.Unblock(now) })
}

func (s *CardService) Cancel(ctx context.Context, id, reason string) (*domain.Card, error) {
	return s.mutate(ctx, id, "card_cancel", func(c *domain.Card, now time.Time) error { return c.Cancel(reason, now) })
}

func (s *CardService) Report(ctx context.Context, id string, cmd ReportCommand) (*domain.Card, *IssuedCard, error) {
	current, err := s.cards.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	var (
		card        *domain.Card
		replacement *IssuedCard
	)
	err = s.locker.WithLock(ctx, accountLockKey(current.AccountID), func(ctx context.Context) error {
		return s.locker.WithLock(ctx, cardLockKey(id), func(ctx context.Context) error {
			c, err := s.cards.Get(ctx, id)
			if err != nil {
				return err
			}
			if err := c.Report(cmd.Reason, cmd.Note, s.clock.Now()); err != nil {
				return err
			}
			if cmd.Reissue {
				account, err := s.accounts.Get(ctx, c.AccountID)
				if err != nil {
					return err
				}
				if account.Status != domain.AccountNormal {
					return domain.Conflict("cannot reissue on a %s account", account.Status)
				}
				program, err := s.catalog.Get(c.ProgramCode)
				if err != nil {
					return err
				}
				replacement, err = s.issue(ctx, program, account, c.Type, c.CardholderName, c)
				if err != nil {
					return err
				}
				c.ReplacedByCardID = replacement.Card.ID
			}
			card = c
			return s.cards.Update(ctx, c)
		})
	})
	if err != nil {
		return nil, nil, err
	}
	s.metrics.ObserveLifecycle("card_report_" + string(cmd.Reason))
	if replacement != nil {
		s.metrics.ObserveIssued(replacement.Card.ProgramCode, replacement.Card.Type)
	}
	return card, replacement, nil
}

func (s *CardService) UpdateLimits(ctx context.Context, id string, limits domain.Limits) (*domain.Card, error) {
	return s.mutate(ctx, id, "card_limits", func(c *domain.Card, now time.Time) error {
		p, err := s.catalog.Get(c.ProgramCode)
		if err != nil {
			return err
		}
		return c.SetLimits(limits, p.MaxLimits, now)
	})
}

func (s *CardService) UpdateControls(ctx context.Context, id string, controls domain.Controls) (*domain.Card, error) {
	return s.mutate(ctx, id, "card_controls", func(c *domain.Card, now time.Time) error {
		p, err := s.catalog.Get(c.ProgramCode)
		if err != nil {
			return err
		}
		return c.SetControls(controls, p.ControlsFor(c.Type), now)
	})
}

func (s *CardService) SetPIN(ctx context.Context, id, pin string) (*domain.Card, error) {
	if err := domain.ValidatePIN(pin); err != nil {
		return nil, err
	}
	hash, err := s.security.HashPIN(pin)
	if err != nil {
		return nil, err
	}
	return s.mutate(ctx, id, "card_set_pin", func(c *domain.Card, now time.Time) error {
		if c.Type != domain.CardPlastic {
			return domain.Invalid("only PLASTIC cards have a PIN")
		}
		return c.SetPIN(hash, now)
	})
}

func (s *CardService) mutate(ctx context.Context, id, event string, fn func(c *domain.Card, now time.Time) error) (*domain.Card, error) {
	var out *domain.Card
	err := s.locker.WithLock(ctx, cardLockKey(id), func(ctx context.Context) error {
		c, err := s.cards.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(c, s.clock.Now()); err != nil {
			return err
		}
		out = c
		return s.cards.Update(ctx, c)
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveLifecycle(event)
	return out, nil
}

func (s *CardService) issue(ctx context.Context, p domain.Program, account *domain.Account, t domain.CardType, name string, replaces *domain.Card) (*IssuedCard, error) {
	now := s.clock.Now()
	for attempt := 0; attempt < maxPANAttempts; attempt++ {
		pan, err := domain.GeneratePAN(p.BIN, s.random)
		if err != nil {
			return nil, err
		}
		cipher, err := s.vault.EncryptPAN(pan)
		if err != nil {
			return nil, err
		}
		id, err := newID(s.random)
		if err != nil {
			return nil, err
		}
		card := &domain.Card{
			ID:             id,
			AccountID:      account.ID,
			CustomerID:     account.CustomerID,
			ProgramCode:    p.Code,
			Type:           t,
			CardholderName: name,
			PANHash:        s.vault.HashPAN(pan),
			PANCipher:      cipher,
			BIN:            pan[:domain.BINLength],
			Last4:          pan[len(pan)-4:],
			Expiry:         domain.NewExpiry(now, p.ValidityYears),
			ServiceCode:    p.ServiceCode(t),
			Status:         domain.StatusNormal,
			Limits:         p.DefaultLimits,
			Controls:       p.ControlsFor(t),
			CreatedAt:      now,
			UpdatedAt:      now,
			ActivatedAt:    &now,
		}
		switch t {
		case domain.CardPlastic:
			card.Status = domain.StatusPending
			card.ActivatedAt = nil
		case domain.CardTemporary:
			until := now.Add(p.TemporaryValidity)
			card.ValidUntil = &until
			card.Expiry = domain.Expiry{Month: int(until.Month()), Year: until.Year()}
		}
		if replaces != nil {
			card.ReplacesCardID = replaces.ID
			card.Limits = replaces.Limits
			card.Controls = replaces.Controls
		}
		err = s.cards.Create(ctx, card)
		if errors.Is(err, domain.ErrDuplicatePAN) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return &IssuedCard{Card: card, PAN: pan, CVV: s.security.CVV(pan, card.Expiry, card.ServiceCode)}, nil
	}
	return nil, fmt.Errorf("could not generate a unique pan for bin %s after %d attempts", p.BIN, maxPANAttempts)
}

func (s *CardService) cvvMatches(c *domain.Card, cvv string) (bool, error) {
	if cvv == "" {
		return false, nil
	}
	pan, err := s.vault.DecryptPAN(c.PANCipher)
	if err != nil {
		return false, err
	}
	want := s.security.CVV(pan, c.Expiry, c.ServiceCode)
	return subtle.ConstantTimeCompare([]byte(cvv), []byte(want)) == 1, nil
}
