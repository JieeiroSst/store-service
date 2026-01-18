package domain

import "time"

type CircleWallet struct {
	ID          string      `json:"id"`
	WalletSetID string      `json:"walletSetId"`
	CustodyType CustodyType `json:"custodyType"`
	Address     string      `json:"address"`
	Blockchain  string      `json:"blockchain"`
	State       WalletState `json:"state"`
	AccountType AccountType `json:"accountType"`
	CreateDate  time.Time   `json:"createDate"`
	UpdateDate  time.Time   `json:"updateDate"`
}

type CustodyType string

const (
	CustodyTypeDeveloper CustodyType = "DEVELOPER"
	CustodyTypeEndUser   CustodyType = "ENDUSER"
)

type WalletState string

const (
	WalletStateLive     WalletState = "LIVE"
	WalletStateFrozen   WalletState = "FROZEN"
	WalletStateInactive WalletState = "INACTIVE"
)

type AccountType string

const (
	AccountTypeSCA AccountType = "SCA" // Smart Contract Account
	AccountTypeEOA AccountType = "EOA" // Externally Owned Account
)

type CircleWalletSet struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	CustodyType CustodyType `json:"custodyType"`
	CreateDate  time.Time   `json:"createDate"`
	UpdateDate  time.Time   `json:"updateDate"`
}

type CreateWalletRequest struct {
	WalletSetID string           `json:"walletSetId"`
	Blockchains []string         `json:"blockchains"`
	Count       int              `json:"count"`
	AccountType AccountType      `json:"accountType,omitempty"`
	Metadata    []WalletMetadata `json:"metadata,omitempty"`
}

type WalletMetadata struct {
	Name  string `json:"name"`
	RefID string `json:"refId"`
}

type UpdateWalletRequest struct {
	Name  string      `json:"name,omitempty"`
	RefID string      `json:"refId,omitempty"`
	State WalletState `json:"state,omitempty"`
}

type CircleBalance struct {
	TokenID    string    `json:"tokenId"`
	Token      Token     `json:"token"`
	Amount     string    `json:"amount"`
	UpdateDate time.Time `json:"updateDate"`
}

type NFTBalance struct {
	TokenID    string    `json:"tokenId"`
	Name       string    `json:"name"`
	Standard   string    `json:"standard"`
	Blockchain string    `json:"blockchain"`
	TokenCount string    `json:"tokenCount"`
	UpdateDate time.Time `json:"updateDate"`
}
