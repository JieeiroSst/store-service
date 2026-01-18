package domain

import "time"

// CircleTransaction represents a Circle transaction
type CircleTransaction struct {
	ID                 string                 `json:"id"`
	Blockchain         string                 `json:"blockchain"`
	TokenID            string                 `json:"tokenId"`
	WalletID           string                 `json:"walletId"`
	SourceAddress      string                 `json:"sourceAddress"`
	DestinationAddress string                 `json:"destinationAddress"`
	TransactionType    TransactionType        `json:"transactionType"`
	CustodyType        CustodyType            `json:"custodyType"`
	State              CircleTransactionState `json:"state"`
	Amounts            []string               `json:"amounts"`
	NFTTokenIDs        []string               `json:"nftTokenIds,omitempty"`
	TxHash             string                 `json:"txHash"`
	BlockHash          string                 `json:"blockHash"`
	BlockHeight        int64                  `json:"blockHeight"`
	NetworkFee         string                 `json:"networkFee"`
	FirstConfirmDate   *time.Time             `json:"firstConfirmDate,omitempty"`
	Operation          Operation              `json:"operation"`
	AbiParameters      []ABIParameter         `json:"abiParameters,omitempty"`
	CreateDate         time.Time              `json:"createDate"`
	UpdateDate         time.Time              `json:"updateDate"`
	UserID             string                 `json:"userId,omitempty"`
	ErrorReason        string                 `json:"errorReason,omitempty"`
}

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeInbound  TransactionType = "INBOUND"
	TransactionTypeOutbound TransactionType = "OUTBOUND"
)

// CircleTransactionState represents Circle transaction states
type CircleTransactionState string

const (
	CircleTxStatePendingRiskScreening CircleTransactionState = "PENDING_RISK_SCREENING"
	CircleTxStateDenied               CircleTransactionState = "DENIED"
	CircleTxStateQueued               CircleTransactionState = "QUEUED"
	CircleTxStateSent                 CircleTransactionState = "SENT"
	CircleTxStateConfirmed            CircleTransactionState = "CONFIRMED"
	CircleTxStateComplete             CircleTransactionState = "COMPLETE"
	CircleTxStateFailed               CircleTransactionState = "FAILED"
	CircleTxStateCancelled            CircleTransactionState = "CANCELLED"
)

// Operation represents the transaction operation
type Operation string

const (
	OperationTransfer           Operation = "TRANSFER"
	OperationContractExecution  Operation = "CONTRACT_EXECUTION"
	OperationContractDeployment Operation = "CONTRACT_DEPLOYMENT"
)

// ABIParameter represents ABI parameters for contract execution
type ABIParameter struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// TransferRequest represents a transfer request
type TransferRequest struct {
	WalletID               string            `json:"walletId"`
	DestinationAddress     string            `json:"destinationAddress"`
	TokenID                string            `json:"tokenId"`
	Amount                 string            `json:"amount"`
	Fee                    *FeeConfiguration `json:"fee,omitempty"`
	IdempotencyKey         string            `json:"idempotencyKey"`
	EntitySecretCiphertext string            `json:"entitySecretCiphertext,omitempty"`
}

// NFTTransferRequest represents an NFT transfer request
type NFTTransferRequest struct {
	WalletID               string            `json:"walletId"`
	DestinationAddress     string            `json:"destinationAddress"`
	TokenID                string            `json:"tokenId"`
	NFTTokenIDs            []string          `json:"nftTokenIds"`
	Fee                    *FeeConfiguration `json:"fee,omitempty"`
	IdempotencyKey         string            `json:"idempotencyKey"`
	EntitySecretCiphertext string            `json:"entitySecretCiphertext,omitempty"`
}

// ContractExecutionRequest represents a contract execution request
type ContractExecutionRequest struct {
	WalletID               string            `json:"walletId"`
	ContractAddress        string            `json:"contractAddress"`
	ABIParameters          []ABIParameter    `json:"abiParameters"`
	Amount                 string            `json:"amount,omitempty"`
	Fee                    *FeeConfiguration `json:"fee,omitempty"`
	IdempotencyKey         string            `json:"idempotencyKey"`
	EntitySecretCiphertext string            `json:"entitySecretCiphertext,omitempty"`
}

// FeeConfiguration represents transaction fee configuration
type FeeConfiguration struct {
	Type   FeeType   `json:"type"`
	Config FeeConfig `json:"config"`
}

// FeeType represents fee type
type FeeType string

const (
	FeeTypeLevel  FeeType = "level"
	FeeTypeCustom FeeType = "custom"
)

// FeeConfig represents fee configuration
type FeeConfig struct {
	FeeLevel    string `json:"feeLevel,omitempty"` // LOW, MEDIUM, HIGH
	MaxFee      string `json:"maxFee,omitempty"`
	PriorityFee string `json:"priorityFee,omitempty"`
	GasLimit    string `json:"gasLimit,omitempty"`
	GasPrice    string `json:"gasPrice,omitempty"`
}

// FeeEstimate represents fee estimation
type FeeEstimate struct {
	Low    FeeLevel `json:"low"`
	Medium FeeLevel `json:"medium"`
	High   FeeLevel `json:"high"`
}

// FeeLevel represents a fee level estimate
type FeeLevel struct {
	MaxFee      string `json:"maxFee"`
	GasLimit    string `json:"gasLimit"`
	BaseFee     string `json:"baseFee"`
	PriorityFee string `json:"priorityFee"`
}

// AccelerateTransactionRequest represents request to accelerate transaction
type AccelerateTransactionRequest struct {
	Fee FeeConfiguration `json:"fee"`
}

// CancelTransactionRequest represents request to cancel transaction
type CancelTransactionRequest struct {
	Fee FeeConfiguration `json:"fee"`
}

// TransactionFilter represents filter for listing transactions
type TransactionFilter struct {
	Blockchain         string                 `json:"blockchain,omitempty"`
	CustodyType        CustodyType            `json:"custodyType,omitempty"`
	DestinationAddress string                 `json:"destinationAddress,omitempty"`
	IncludeAll         bool                   `json:"includeAll,omitempty"`
	Operation          Operation              `json:"operation,omitempty"`
	State              CircleTransactionState `json:"state,omitempty"`
	TokenID            string                 `json:"tokenId,omitempty"`
	TxHash             string                 `json:"txHash,omitempty"`
	TxType             TransactionType        `json:"txType,omitempty"`
	UserID             string                 `json:"userId,omitempty"`
	WalletIDs          []string               `json:"walletIds,omitempty"`
	From               *time.Time             `json:"from,omitempty"`
	To                 *time.Time             `json:"to,omitempty"`
	PageBefore         string                 `json:"pageBefore,omitempty"`
	PageAfter          string                 `json:"pageAfter,omitempty"`
	PageSize           int                    `json:"pageSize,omitempty"`
}
