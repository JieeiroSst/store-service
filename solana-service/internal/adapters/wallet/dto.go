package wallet

import "time"

// API request/response structures
type createWalletRequest struct {
	AccountType    string            `json:"accountType"`
	Blockchain     string            `json:"blockchain"`
	WalletSetID    string            `json:"walletSetId"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey"`
}

type walletResponse struct {
	Data struct {
		Wallet circleWalletData `json:"wallet"`
	} `json:"data"`
}

type walletsResponse struct {
	Data struct {
		Wallets []circleWalletData `json:"wallets"`
	} `json:"data"`
}

type circleWalletData struct {
	ID          string    `json:"id"`
	State       string    `json:"state"`
	WalletSetID string    `json:"walletSetId"`
	CustodyType string    `json:"custodyType"`
	Address     string    `json:"address"`
	Blockchain  string    `json:"blockchain"`
	AccountType string    `json:"accountType"`
	UpdateDate  time.Time `json:"updateDate"`
	CreateDate  time.Time `json:"createDate"`
}

type createWalletSetRequest struct {
	Name string `json:"name"`
}

type walletSetResponse struct {
	Data struct {
		WalletSet circleWalletSetData `json:"walletSet"`
	} `json:"data"`
}

type circleWalletSetData struct {
	ID          string    `json:"id"`
	CustodyType string    `json:"custodyType"`
	Name        string    `json:"name"`
	UpdateDate  time.Time `json:"updateDate"`
	CreateDate  time.Time `json:"createDate"`
}

type createTransferRequest struct {
	WalletID           string              `json:"walletId"`
	DestinationAddress string              `json:"destinationAddress"`
	TokenID            string              `json:"tokenId"`
	Amount             string              `json:"amount"`
	Fee                *transferFeeRequest `json:"fee,omitempty"`
	IdempotencyKey     string              `json:"idempotencyKey"`
}

type transferFeeRequest struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

type transferResponse struct {
	Data struct {
		Transfer circleTransferData `json:"transfer"`
	} `json:"data"`
}

type circleTransferData struct {
	ID                 string    `json:"id"`
	State              string    `json:"state"`
	WalletID           string    `json:"walletId"`
	DestinationAddress string    `json:"destinationAddress"`
	TokenID            string    `json:"tokenId"`
	Amount             string    `json:"amount"`
	TransactionHash    string    `json:"transactionHash"`
	CreateDate         time.Time `json:"createDate"`
}

type balanceResponse struct {
	Data struct {
		TokenBalances []tokenBalance `json:"tokenBalances"`
	} `json:"data"`
}

type tokenBalance struct {
	Token  tokenInfo `json:"token"`
	Amount string    `json:"amount"`
}

type tokenInfo struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}
