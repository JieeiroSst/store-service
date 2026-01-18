package domain

import "errors"

var (
	// Solana errors
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidTransaction  = errors.New("invalid transaction")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrProgramNotFound     = errors.New("program not found")
	ErrInvalidPDA          = errors.New("invalid program derived address")
	ErrCPIFailed           = errors.New("cross program invocation failed")
	ErrInvalidPublicKey    = errors.New("invalid public key")

	// Circle errors
	ErrCircleAPIFailed         = errors.New("circle api request failed")
	ErrWalletNotFound          = errors.New("circle wallet not found")
	ErrWalletSetNotFound       = errors.New("circle wallet set not found")
	ErrInvalidAmount           = errors.New("invalid amount")
	ErrTransferFailed          = errors.New("circle transfer failed")
	ErrInsufficientCircleFunds = errors.New("insufficient circle wallet funds")
	ErrTokenNotFound           = errors.New("token not found")
	ErrInvalidToken            = errors.New("invalid token")
	ErrNFTTransferFailed       = errors.New("nft transfer failed")
	ErrContractExecutionFailed = errors.New("contract execution failed")
	ErrFeeEstimationFailed     = errors.New("fee estimation failed")
	ErrChallengeRequired       = errors.New("user challenge required")

	// Bridge errors
	ErrBridgeTransferFailed = errors.New("bridge transfer failed")
	ErrInvalidDirection     = errors.New("invalid bridge direction")
	ErrBridgeNotFound       = errors.New("bridge transfer not found")
)
