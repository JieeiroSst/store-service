package grpc

import (
	"errors"

	grpcapi "github.com/JIeeiroSst/lib-gateway/account-transaction-service/gateway/account-transaction-service"
	"github.com/Jieeirosst/account-transaction-service/common"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func accountToPB(a model.Account) *grpcapi.Account {
	return &grpcapi.Account{
		Id:          a.ID,
		FirstName:   a.FirstName,
		LastName:    a.LastName,
		DateCreated: timestamppb.New(a.DateCreated),
	}
}

func accountsToPB(accounts []model.Account) []*grpcapi.Account {
	pbAccounts := make([]*grpcapi.Account, 0, len(accounts))
	for _, a := range accounts {
		pbAccounts = append(pbAccounts, accountToPB(a))
	}
	return pbAccounts
}

func transactionToPB(t model.Transaction) *grpcapi.Transaction {
	pb := &grpcapi.Transaction{
		Id:          t.ID,
		Type:        string(t.Type),
		Amount:      t.Amount,
		DateCreated: timestamppb.New(t.DateCreated),
	}

	switch t.Type {
	case model.TransactionTypeDeposit:
		pb.TransactionDetails = &grpcapi.Transaction_Deposit{
			Deposit: &grpcapi.DepositTransactionDetails{AccountId: t.AccountID},
		}
	case model.TransactionTypeWithdrawal:
		pb.TransactionDetails = &grpcapi.Transaction_Withdrawal{
			Withdrawal: &grpcapi.WithdrawalTransactionDetails{AccountId: t.AccountID},
		}
	case model.TransactionTypeTransfer:
		pb.TransactionDetails = &grpcapi.Transaction_Transfer{
			Transfer: &grpcapi.AccountToAccountTransactionDetails{
				SenderId:   t.SenderID,
				ReceiverId: t.ReceiverID,
			},
		}
	case model.TransactionTypePayment:
		pb.TransactionDetails = &grpcapi.Transaction_Payment{
			Payment: &grpcapi.PaymentForServiceTransactionDetails{
				AccountId:   t.AccountID,
				ServiceName: t.ServiceName,
			},
		}
	}

	return pb
}

func transactionsToPB(txs []model.Transaction) []*grpcapi.Transaction {
	pbTxs := make([]*grpcapi.Transaction, 0, len(txs))
	for _, t := range txs {
		pbTxs = append(pbTxs, transactionToPB(t))
	}
	return pbTxs
}

// mapError translates domain/application errors into gRPC status errors so
// gateway clients see the correct HTTP status instead of a generic 500.
func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, common.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, common.ErrInvalidAmount),
		errors.Is(err, common.ErrSameAccount),
		errors.Is(err, common.ErrInvalidTransactionType):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
