package application

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const paymentRequestKeyPrefix = "payreq:"

type paymentRequestService struct {
	core
}

func NewPaymentRequestService(d Deps) port.PaymentRequestService {
	return &paymentRequestService{core: newCore(d)}
}

func (s *paymentRequestService) Create(ctx context.Context, cmd port.CreatePaymentRequestCommand) (*model.PaymentRequest, error) {
	w, err := s.ownedWallet(ctx, cmd.RequesterWalletID, cmd.RequesterUserID)
	if err != nil {
		return nil, err
	}
	if err := w.EnsureActive(); err != nil {
		return nil, err
	}
	pr, err := model.NewPaymentRequest(w, cmd.PayerUserID, cmd.Amount, cmd.Note, s.PaymentRequest.TTL, s.now())
	if err != nil {
		return nil, err
	}
	if err := s.PaymentRequests.Create(ctx, pr); err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *paymentRequestService) Get(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error) {
	pr, err := s.PaymentRequests.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	pr.Expire(s.now())
	return pr, nil
}

func (s *paymentRequestService) List(ctx context.Context, filter model.PaymentRequestFilter) ([]*model.PaymentRequest, int64, error) {
	filter.Role = model.PaymentRequestRole(strings.ToLower(string(filter.Role)))
	if filter.Role == "" {
		filter.Role = model.RoleIncoming
	}
	if filter.Role != model.RoleIncoming && filter.Role != model.RoleOutgoing {
		return nil, 0, model.Invalid("role must be incoming or outgoing")
	}
	filter.Status = model.PaymentRequestStatus(strings.ToUpper(string(filter.Status)))
	filter.Page, filter.PageSize = model.NormalizePage(filter.Page, filter.PageSize)
	items, total, err := s.PaymentRequests.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	now := s.now()
	for _, pr := range items {
		pr.Expire(now)
	}
	return items, total, nil
}

func (s *paymentRequestService) Pay(ctx context.Context, cmd port.PayPaymentRequestCommand) (*model.PaymentRequest, error) {
	pr, err := s.PaymentRequests.GetByID(ctx, cmd.RequestID)
	if err != nil {
		return nil, err
	}
	if pr.Status == model.PaymentRequestPaid && pr.PayerUserID == cmd.UserID {
		return pr, nil
	}
	if err := pr.EnsurePayable(cmd.UserID, s.now()); err != nil {
		return nil, err
	}
	payer, err := s.ownedWallet(ctx, cmd.WalletID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if payer.Currency != pr.Currency {
		return nil, model.ErrCurrencyMismatch
	}
	if err := s.verifyPin(ctx, payer.ID, cmd.UserID, cmd.Pin); err != nil {
		return nil, err
	}

	err = s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		locked, err := s.PaymentRequests.GetByIDForUpdate(ctx, cmd.RequestID)
		if err != nil {
			return err
		}
		now := s.now()
		if err := locked.EnsurePayable(cmd.UserID, now); err != nil {
			return err
		}
		description := "Payment request"
		if locked.Note != "" {
			description += ": " + locked.Note
		}
		tx, err := s.transferWithinTx(ctx, transferInput{
			idempotencyKey: paymentRequestKeyPrefix + locked.ID.String(),
			userID:         cmd.UserID,
			fromWalletID:   payer.ID,
			toWalletID:     locked.RequesterWalletID,
			amount:         locked.Amount,
			description:    description,
			metadata:       model.Metadata{"payment_request_id": locked.ID.String()},
		})
		if err != nil {
			return err
		}
		locked.MarkPaid(payer.ID, tx.ID, now)
		pr = locked
		return s.PaymentRequests.Update(ctx, locked)
	})
	if err != nil {
		return nil, err
	}
	s.invalidateWallets(ctx, payer.ID, pr.RequesterWalletID)
	s.Logger.Info("payment request paid", zap.String("request_id", pr.ID.String()), zap.String("tx_id", pr.TransactionID.UUID.String()))
	return pr, nil
}

func (s *paymentRequestService) Decline(ctx context.Context, id, userID uuid.UUID) (*model.PaymentRequest, error) {
	return s.transition(ctx, id, func(pr *model.PaymentRequest) error { return pr.Decline(userID, s.now()) })
}

func (s *paymentRequestService) Cancel(ctx context.Context, id, userID uuid.UUID) (*model.PaymentRequest, error) {
	return s.transition(ctx, id, func(pr *model.PaymentRequest) error { return pr.Cancel(userID, s.now()) })
}

func (s *paymentRequestService) transition(ctx context.Context, id uuid.UUID, fn func(*model.PaymentRequest) error) (*model.PaymentRequest, error) {
	var out *model.PaymentRequest
	err := s.Tx.WithinTx(ctx, func(ctx context.Context) error {
		pr, err := s.PaymentRequests.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if err := fn(pr); err != nil {
			return err
		}
		out = pr
		return s.PaymentRequests.Update(ctx, pr)
	})
	return out, err
}
