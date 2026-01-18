package domain

import "time"

type Challenge struct {
	ID             string        `json:"id"`
	CorrelationIDs []string      `json:"correlationIds"`
	Type           ChallengeType `json:"type"`
}

type ChallengeType string

const (
	ChallengeTypeSetPin      ChallengeType = "SET_PIN"
	ChallengeTypeRestorePin  ChallengeType = "RESTORE_PIN"
	ChallengeTypeTransaction ChallengeType = "TRANSACTION"
)

type BridgeTransfer struct {
	ID                string          `json:"id"`
	SolanaAddress     string          `json:"solanaAddress"`
	CircleWalletID    string          `json:"circleWalletId"`
	Direction         BridgeDirection `json:"direction"`
	Amount            string          `json:"amount"`
	TokenID           string          `json:"tokenId"`
	Status            BridgeStatus    `json:"status"`
	SolanaTxSignature string          `json:"solanaTxSignature,omitempty"`
	CircleTxID        string          `json:"circleTxId,omitempty"`
	CreateDate        time.Time       `json:"createDate"`
	CompleteDate      *time.Time      `json:"completeDate,omitempty"`
	ErrorMessage      string          `json:"errorMessage,omitempty"`
}

type BridgeDirection string

const (
	DirectionSolanaToCircle BridgeDirection = "SOLANA_TO_CIRCLE"
	DirectionCircleToSolana BridgeDirection = "CIRCLE_TO_SOLANA"
)

type BridgeStatus string

const (
	BridgeStatusInitiated BridgeStatus = "INITIATED"
	BridgeStatusPending   BridgeStatus = "PENDING"
	BridgeStatusCompleted BridgeStatus = "COMPLETED"
	BridgeStatusFailed    BridgeStatus = "FAILED"
)
