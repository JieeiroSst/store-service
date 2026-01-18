package domain

import "github.com/gagliardetto/solana-go"

type Transaction struct {
	Signature    string
	BlockTime    *int64
	Slot         uint64
	Fee          uint64
	Status       TransactionStatus
	Instructions []Instruction
}

type TransactionStatus string

const (
	StatusPending   TransactionStatus = "pending"
	StatusConfirmed TransactionStatus = "confirmed"
	StatusFinalized TransactionStatus = "finalized"
	StatusFailed    TransactionStatus = "failed"
)

type Instruction struct {
	ProgramID string
	Accounts  []string
	Data      []byte
}

type TransactionRequest struct {
	From            solana.PublicKey
	To              solana.PublicKey
	Amount          uint64
	RecentBlockhash solana.Hash
}

type FeeInfo struct {
	Fee                  uint64
	FeeCalculator        FeeCalculator
	LastValidBlockHeight uint64
}

type FeeCalculator struct {
	LamportsPerSignature uint64
}
