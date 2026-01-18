package domain

import (
	"github.com/gagliardetto/solana-go"
)

type Program struct {
	ID         solana.PublicKey
	Name       string
	Executable bool
	Owner      solana.PublicKey
}

type PDA struct {
	Address solana.PublicKey
	Bump    uint8
	Seeds   [][]byte
}

type CPIRequest struct {
	CallerProgram solana.PublicKey
	TargetProgram solana.PublicKey
	Instruction   []byte
	Accounts      []solana.AccountMeta
}
