package http

type TransferRequest struct {
	From   string `json:"from" binding:"required"`
	To     string `json:"to" binding:"required"`
	Amount uint64 `json:"amount" binding:"required"`
}

type PDARequest struct {
	ProgramID  string `json:"program_id" binding:"required"`
	UserPubkey string `json:"user_pubkey" binding:"required"`
	Identifier string `json:"identifier" binding:"required"`
}

type CreateTransferRequest struct {
	WalletID           string              `json:"wallet_id" binding:"required"`
	DestinationAddress string              `json:"destination_address" binding:"required"`
	TokenID            string              `json:"token_id" binding:"required"`
	Amount             string              `json:"amount" binding:"required"`
	Fee                *TransferFeeRequest `json:"fee"`
	IdempotencyKey     string              `json:"idempotency_key"`
}

type TransferFeeRequest struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}
