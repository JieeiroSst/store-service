package kafka

import (
	"encoding/json"
	"fmt"
)

type TransactionEvent struct {
	ExternalRef     string `json:"external_ref"`
	AccountID       int    `json:"account_id"`
	TransactionType string `json:"transaction_type"`
	Amount          uint   `json:"amount"`
	TransactionDate int    `json:"transaction_date"`
}

func parseTransactionEvent(data []byte) (*TransactionEvent, error) {
	var evt TransactionEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		return nil, fmt.Errorf("invalid transaction event payload: %w", err)
	}
	if evt.ExternalRef == "" {
		return nil, fmt.Errorf("external_ref is required")
	}
	if evt.AccountID == 0 {
		return nil, fmt.Errorf("account_id is required")
	}
	if evt.TransactionType == "" {
		return nil, fmt.Errorf("transaction_type is required")
	}
	return &evt, nil
}
