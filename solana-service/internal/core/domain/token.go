package domain

import (
	"time"
)

type Token struct {
	ID           string        `json:"id"`
	Blockchain   string        `json:"blockchain"`
	Name         string        `json:"name"`
	Symbol       string        `json:"symbol"`
	Decimals     int           `json:"decimals"`
	IsNative     bool          `json:"isNative"`
	TokenAddress string        `json:"tokenAddress,omitempty"`
	Standard     TokenStandard `json:"standard,omitempty"`
	CreateDate   time.Time     `json:"createDate"`
	UpdateDate   time.Time     `json:"updateDate"`
}

type TokenStandard string

const (
	TokenStandardERC20   TokenStandard = "ERC20"
	TokenStandardERC721  TokenStandard = "ERC721"
	TokenStandardERC1155 TokenStandard = "ERC1155"
	TokenStandardSPL     TokenStandard = "SPL"
)

type NFTToken struct {
	ID           string    `json:"id"`
	Blockchain   string    `json:"blockchain"`
	TokenAddress string    `json:"tokenAddress"`
	Standard     string    `json:"standard"`
	Name         string    `json:"name"`
	Symbol       string    `json:"symbol"`
	CreateDate   time.Time `json:"createDate"`
}

type TokenMetadata struct {
	Name         string        `json:"name"`
	Symbol       string        `json:"symbol"`
	Decimals     int           `json:"decimals"`
	Blockchain   string        `json:"blockchain"`
	TokenAddress string        `json:"tokenAddress"`
	Standard     TokenStandard `json:"standard"`
}
