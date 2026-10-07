package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
	"github.com/JIeeiroSst/auth-service/internal/port"
)

type Settings struct {
	Policy    domain.Policy
	ExposeOTP bool
}

type AuthenticationService struct {
	challenges port.ChallengeRepository
	locker     port.Locker
	cards      port.CardDirectory
	sessions   port.SessionValidator
	users      port.UserDirectory
	otp        port.OTPProvider
	sender     port.OTPSender
	clock      port.Clock
	metrics    port.Metrics
	settings   Settings
	random     io.Reader
}

func NewAuthenticationService(settings Settings, challenges port.ChallengeRepository, locker port.Locker, cards port.CardDirectory, sessions port.SessionValidator, users port.UserDirectory, otp port.OTPProvider, sender port.OTPSender, clock port.Clock, metrics port.Metrics) *AuthenticationService {
	return &AuthenticationService{
		challenges: challenges,
		locker:     locker,
		cards:      cards,
		sessions:   sessions,
		users:      users,
		otp:        otp,
		sender:     sender,
		clock:      clock,
		metrics:    metrics,
		settings:   settings,
		random:     rand.Reader,
	}
}

type Initiated struct {
	Challenge *domain.Challenge
	OTP       string
}

type ConsumeCommand struct {
	CardID   string
	Amount   int64
	Currency string
}

func (s *AuthenticationService) Initiate(ctx context.Context, req domain.InitiateRequest) (*Initiated, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	card, err := s.cards.ResolvePAN(ctx, req.PAN, req.Expiry)
	if err != nil {
		return nil, err
	}
	if card.UserID <= 0 {
		return nil, domain.Conflict("card has no cardholder enrolled for payment authentication")
	}
	username, err := s.users.Username(ctx, card.UserID)
	if err != nil {
		return nil, err
	}
	id, err := newID(s.random)
	if err != nil {
		return nil, err
	}
	code, otpExpiresAt, err := s.otp.Issue(ctx, username)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	c, err := domain.NewChallenge(id, card, username, req, otpExpiresAt, s.settings.Policy, now)
	if err != nil {
		return nil, err
	}
	if err := s.challenges.Create(ctx, c); err != nil {
		return nil, err
	}
	if err := s.deliver(ctx, c, code); err != nil {
		return nil, err
	}
	s.metrics.ObserveChallenge("initiated")
	out := &Initiated{Challenge: c}
	if s.settings.ExposeOTP {
		out.OTP = code
	}
	return out, nil
}

func (s *AuthenticationService) deliver(ctx context.Context, c *domain.Challenge, code string) error {
	err := s.sender.Send(ctx, port.OTPMessage{
		UserID: c.UserID, ChallengeID: c.ID, OTP: code, MaskedPAN: c.MaskedPAN,
		Amount: c.Amount, Currency: c.Currency, Merchant: c.Merchant, ExpiresAt: c.OTPExpiresAt,
	})
	if err != nil {
		return fmt.Errorf("%w: deliver otp: %v", domain.ErrUnavailable, err)
	}
	return nil
}

func (s *AuthenticationService) Resend(ctx context.Context, id, sessionToken string) (*Initiated, error) {
	userID, err := s.sessions.Validate(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	var (
		out  *domain.Challenge
		code string
	)
	err = s.locker.WithLock(ctx, challengeLockKey(id), func(ctx context.Context) error {
		c, err := s.challenges.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := c.CanResend(userID, s.clock.Now(), s.settings.Policy); err != nil {
			if errors.Is(err, domain.ErrExpired) {
				_ = s.challenges.Update(ctx, c)
			}
			return err
		}
		var expiresAt time.Time
		code, expiresAt, err = s.otp.Issue(ctx, c.Username)
		if err != nil {
			return err
		}
		c.Resent(expiresAt, s.clock.Now())
		if err := s.challenges.Update(ctx, c); err != nil {
			return err
		}
		out = c
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.deliver(ctx, out, code); err != nil {
		return nil, err
	}
	s.metrics.ObserveChallenge("resent")
	res := &Initiated{Challenge: out}
	if s.settings.ExposeOTP {
		res.OTP = code
	}
	return res, nil
}

func (s *AuthenticationService) Get(ctx context.Context, id string) (*domain.Challenge, error) {
	c, err := s.challenges.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Refresh(s.clock.Now())
	return c, nil
}

func (s *AuthenticationService) ListPending(ctx context.Context, sessionToken string) ([]*domain.Challenge, error) {
	userID, err := s.sessions.Validate(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	return s.challenges.ListPendingByUser(ctx, userID, s.clock.Now())
}

func (s *AuthenticationService) Verify(ctx context.Context, id, sessionToken, otp string) (*domain.Challenge, error) {
	otp = strings.TrimSpace(otp)
	if len(otp) != domain.OTPLength {
		return nil, domain.Invalid("otp must be %d digits", domain.OTPLength)
	}
	userID, err := s.sessions.Validate(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	var (
		out       *domain.Challenge
		verifyErr error
	)
	err = s.locker.WithLock(ctx, challengeLockKey(id), func(ctx context.Context) error {
		c, err := s.challenges.Get(ctx, id)
		if err != nil {
			return err
		}
		if verifyErr = c.Precheck(userID, s.clock.Now()); verifyErr != nil {
			if errors.Is(verifyErr, domain.ErrExpired) {
				out = c
				return s.challenges.Update(ctx, c)
			}
			return nil
		}
		matches, err := s.otp.Verify(ctx, c.Username, otp)
		if err != nil {
			return err
		}
		verifyErr = c.Verify(userID, matches, s.clock.Now(), s.settings.Policy)
		out = c
		return s.challenges.Update(ctx, c)
	})
	if err != nil {
		return nil, err
	}
	switch {
	case verifyErr == nil:
		s.metrics.ObserveChallenge("authenticated")
	case domain.IsOTPMismatch(verifyErr) && out.Status == domain.StatusFailed:
		s.metrics.ObserveChallenge("failed")
	case domain.IsOTPMismatch(verifyErr):
		s.metrics.ObserveChallenge("otp_mismatch")
	}
	return out, verifyErr
}

func (s *AuthenticationService) Decline(ctx context.Context, id, sessionToken string) (*domain.Challenge, error) {
	userID, err := s.sessions.Validate(ctx, sessionToken)
	if err != nil {
		return nil, err
	}
	var out *domain.Challenge
	err = s.locker.WithLock(ctx, challengeLockKey(id), func(ctx context.Context) error {
		c, err := s.challenges.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := c.Decline(userID, s.clock.Now()); err != nil {
			if errors.Is(err, domain.ErrExpired) {
				_ = s.challenges.Update(ctx, c)
			}
			return err
		}
		out = c
		return s.challenges.Update(ctx, c)
	})
	if err != nil {
		return nil, err
	}
	s.metrics.ObserveChallenge("declined")
	return out, nil
}

func (s *AuthenticationService) Consume(ctx context.Context, id string, cmd ConsumeCommand) (*domain.Challenge, error) {
	var out *domain.Challenge
	err := s.locker.WithLock(ctx, challengeLockKey(id), func(ctx context.Context) error {
		c, err := s.challenges.Get(ctx, id)
		if err != nil {
			return err
		}
		if err := c.Consume(cmd.CardID, cmd.Amount, cmd.Currency, s.clock.Now()); err != nil {
			if errors.Is(err, domain.ErrExpired) {
				_ = s.challenges.Update(ctx, c)
			}
			return err
		}
		out = c
		return s.challenges.Update(ctx, c)
	})
	if err != nil {
		s.metrics.ObserveChallenge("consume_rejected")
		return nil, err
	}
	s.metrics.ObserveChallenge("consumed")
	return out, nil
}

func challengeLockKey(id string) string { return "challenge:" + id }

func newID(r io.Reader) (string, error) {
	var b [16]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
