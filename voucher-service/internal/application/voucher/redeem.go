package voucher

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"go.uber.org/zap"
)

const (
	redeemLockTTL        = 5 * time.Second
	idempotencyRecordTTL = 24 * time.Hour
	providerCallTimeout  = 3 * time.Second
)

func lockKeyForRedeem(id shared.VoucherID) string {
	return "lock:voucher:redeem:" + id.String()
}

func requestHash(in RedeemVoucherInput) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%s", in.VoucherID.String(), in.Amount.Amount, in.Amount.Currency)))
	return hex.EncodeToString(h[:])
}

func (s *Service) RedeemVoucher(ctx context.Context, in RedeemVoucherInput) (*RedeemVoucherOutput, error) {
	if in.VoucherID.IsZero() || in.IdempotencyKey == "" {
		return nil, shared.ErrValidation
	}

	// Step 2: Redis lock, fail closed.
	release, ok, err := s.locker.Acquire(ctx, lockKeyForRedeem(in.VoucherID), redeemLockTTL)
	if err != nil {
		s.log.Error("redeem: lock backend unavailable", zap.Error(err))
		return nil, ErrLockUnavailable
	}
	if !ok {
		return nil, ErrRedeemInProgress
	}
	defer func() { _ = release(ctx) }()

	// Step 3: idempotency claim.
	reqHash := requestHash(in)
	claimed, err := s.idemp.Claim(ctx, in.IdempotencyKey, reqHash, idempotencyRecordTTL)
	if err != nil {
		return nil, fmt.Errorf("idempotency claim: %w", err)
	}
	if !claimed {
		record, err := s.idemp.Get(ctx, in.IdempotencyKey)
		if err != nil {
			return nil, err
		}
		switch record.Status {
		case idempotency.StatusCompleted:
			var out RedeemVoucherOutput
			if err := json.Unmarshal(record.ResponseBody, &out); err != nil {
				return nil, err
			}
			return &out, nil
		case idempotency.StatusFailed:
			return nil, mapCachedFailure(record)
		default:
			return nil, ErrDuplicateRequestInProgress
		}
	}

	var output *RedeemVoucherOutput
	txErr := s.txManager.WithinTx(ctx, func(ctx context.Context) error {
		// Step 5: row lock.
		v, err := s.repo.FindByIDForUpdate(ctx, in.VoucherID)
		if err != nil {
			return err
		}

		// Step 6: domain validation.
		if v.IsExpired(s.clock.Now()) {
			return voucher.ErrVoucherExpired
		}
		if err := v.ValidatePIN(in.PIN); err != nil {
			return err
		}
		if !v.CanRedeem(s.clock.Now()) {
			return voucher.ErrInvalidTransition
		}

		// Step 7: for API providers, call the merchant before persisting.
		info, err := s.merchantLookup.GetMerchantInfo(ctx, v.MerchantID)
		if err != nil {
			return err
		}
		providerTxnRef := ""
		if info.ProviderType == shared.ProviderTypeAPI {
			provider, err := s.registry.Resolve(shared.ProviderTypeAPI)
			if err != nil {
				return err
			}
			providerCtx, cancel := context.WithTimeout(ctx, providerCallTimeout)
			defer cancel()
			result, err := provider.Redeem(providerCtx, v.Code, in.PIN, in.Amount)
			if err != nil {
				// Any transport-level failure (timeout, network) is treated
				// as transient: the local transition is discarded below and
				// the caller stays free to retry.
				return ErrProviderTimeout
			}
			if !result.Success {
				return ErrProviderRejected
			}
			providerTxnRef = result.ProviderTxnRef
		}

		// Step 6 continued: apply the transition now that any external call
		// has succeeded.
		if err := v.Redeem(in.Amount, providerTxnRef, s.clock.Now()); err != nil {
			return err
		}

		// Step 8: persist with optimistic-lock guard.
		if err := s.repo.Save(ctx, v); err != nil {
			return err
		}

		// Step 9: outbox, same tx.
		for _, evt := range v.PullEvents() {
			outboxEvt, err := outbox.NewEventFromDomain(aggregateType, outboxTopic, evt)
			if err != nil {
				return err
			}
			if err := s.outboxP.Enqueue(ctx, outboxEvt); err != nil {
				return err
			}
		}

		output = &RedeemVoucherOutput{
			VoucherID:      v.ID.String(),
			Status:         string(v.Status),
			RedeemedAmount: in.Amount,
			ProviderTxnRef: providerTxnRef,
		}
		return nil
	})

	if txErr != nil {
		return nil, s.finalizeRedeemFailure(ctx, in.IdempotencyKey, txErr)
	}

	// Step 10: idempotency finalize (best-effort; tx already committed).
	body, _ := json.Marshal(output)
	if err := s.idemp.Complete(ctx, in.IdempotencyKey, 200, body); err != nil {
		s.log.Warn("redeem: idempotency complete failed after commit", zap.Error(err))
	}

	return output, nil
}

func (s *Service) finalizeRedeemFailure(ctx context.Context, key string, err error) error {
	switch {
	case errors.Is(err, ErrProviderTimeout):
		if releaseErr := s.idemp.Release(ctx, key); releaseErr != nil {
			s.log.Error("redeem: idempotency release failed", zap.Error(releaseErr))
		}
		return err
	case errors.Is(err, voucher.ErrInvalidTransition),
		errors.Is(err, voucher.ErrVoucherExpired),
		errors.Is(err, voucher.ErrAlreadyRedeemed),
		errors.Is(err, voucher.ErrInvalidPIN),
		errors.Is(err, ErrProviderRejected):
		if failErr := s.idemp.Fail(ctx, key); failErr != nil {
			s.log.Error("redeem: idempotency fail-cache failed", zap.Error(failErr))
		}
		return err
	default:
		if releaseErr := s.idemp.Release(ctx, key); releaseErr != nil {
			s.log.Error("redeem: idempotency release failed", zap.Error(releaseErr))
		}
		s.log.Error("redeem: transaction failed", zap.Error(err))
		return ErrTransactionFailed
	}
}

func mapCachedFailure(record *idempotency.Record) error {
	return fmt.Errorf("%w: cached failure for idempotency key", shared.ErrDuplicateRequest)
}
