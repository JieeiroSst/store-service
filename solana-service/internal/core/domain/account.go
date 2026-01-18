package domain

import (
	"github.com/gagliardetto/solana-go"
)

type Account struct {
	PublicKey  solana.PublicKey
	Lamports   uint64
	Owner      solana.PublicKey
	Data       []byte
	Executable bool
	RentEpoch  uint64
}

type AccountInfo struct {
	Address    string
	Balance    uint64
	Owner      string
	Executable bool
	RentEpoch  uint64
}
