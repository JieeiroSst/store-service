package model

import (
	"encoding/json"
	"time"
)

const SigningPayloadType = "crm-contract-signature/v1"

type signingPayload struct {
	Type        string `json:"type"`
	ContractID  uint   `json:"contract_id"`
	AccountID   uint   `json:"account_id"`
	SignedBy    string `json:"signed_by"`
	EndDate     string `json:"end_date"`
	SigningTime string `json:"signing_time"`
}

func SigningPayload(c *Contract, signedBy string, endDate, signingTime time.Time) []byte {
	b, _ := json.Marshal(signingPayload{
		Type:        SigningPayloadType,
		ContractID:  c.ID,
		AccountID:   c.AccountID,
		SignedBy:    signedBy,
		EndDate:     endDate.UTC().Format(time.RFC3339),
		SigningTime: signingTime.UTC().Format(time.RFC3339),
	})
	return b
}
