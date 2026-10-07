package domain

import (
	"sort"
	"time"
)

type Mode string

const (
	ModeDebit   Mode = "DEBIT"
	ModeCredit  Mode = "CREDIT"
	ModePrepaid Mode = "PREPAID"
)

type CardType string

const (
	CardPlastic   CardType = "PLASTIC"
	CardVirtual   CardType = "VIRTUAL"
	CardTemporary CardType = "TEMPORARY"
)

func ParseCardType(s string) (CardType, error) {
	switch t := CardType(s); t {
	case CardPlastic, CardVirtual, CardTemporary:
		return t, nil
	}
	return "", Invalid("card type must be PLASTIC, VIRTUAL or TEMPORARY, got %q", s)
}

type Limits struct {
	PerTransaction int64
	Daily          int64
}

func (l Limits) Validate(max Limits) error {
	if l.PerTransaction <= 0 || l.Daily <= 0 {
		return Invalid("limits must be positive")
	}
	if l.PerTransaction > l.Daily {
		return Invalid("per-transaction limit %d exceeds daily limit %d", l.PerTransaction, l.Daily)
	}
	if l.PerTransaction > max.PerTransaction {
		return Invalid("per-transaction limit %d exceeds program maximum %d", l.PerTransaction, max.PerTransaction)
	}
	if l.Daily > max.Daily {
		return Invalid("daily limit %d exceeds program maximum %d", l.Daily, max.Daily)
	}
	return nil
}

type Program struct {
	Code               string
	Name               string
	Mode               Mode
	BIN                string
	Currency           string
	ValidityYears      int
	TemporaryValidity  time.Duration
	CardTypes          []CardType
	MaxActiveVirtual   int
	DefaultCreditLimit int64
	MaxCreditLimit     int64
	DefaultLimits      Limits
	MaxLimits          Limits
	Supported          Controls
	StepUpThreshold    int64
}

func (p Program) Allows(t CardType) bool {
	for _, ct := range p.CardTypes {
		if ct == t {
			return true
		}
	}
	return false
}

func (p Program) ServiceCode(t CardType) string {
	if t == CardPlastic {
		return "201"
	}
	return "101"
}

func (p Program) ControlsFor(t CardType) Controls {
	c := p.Supported
	if t != CardPlastic {
		c.POS, c.Contactless, c.ATM = false, false, false
	}
	return c
}

var Programs = []Program{
	{
		Code:              "DEBIT_CLASSIC",
		Name:              "SARD Debit Classic",
		Mode:              ModeDebit,
		BIN:               "730100",
		Currency:          "VND",
		ValidityYears:     5,
		TemporaryValidity: 24 * time.Hour,
		CardTypes:         []CardType{CardPlastic, CardVirtual, CardTemporary},
		MaxActiveVirtual:  3,
		DefaultLimits:     Limits{PerTransaction: 20_000_000, Daily: 50_000_000},
		MaxLimits:         Limits{PerTransaction: 100_000_000, Daily: 200_000_000},
		Supported:         Controls{POS: true, Contactless: true, Ecommerce: true, ATM: true},
		StepUpThreshold:   2_000_000,
	},
	{
		Code:              "DEBIT_PLATINUM",
		Name:              "SARD Debit Platinum",
		Mode:              ModeDebit,
		BIN:               "730200",
		Currency:          "VND",
		ValidityYears:     5,
		TemporaryValidity: 24 * time.Hour,
		CardTypes:         []CardType{CardPlastic, CardVirtual, CardTemporary},
		MaxActiveVirtual:  5,
		DefaultLimits:     Limits{PerTransaction: 100_000_000, Daily: 300_000_000},
		MaxLimits:         Limits{PerTransaction: 500_000_000, Daily: 1_000_000_000},
		Supported:         Controls{POS: true, Contactless: true, Ecommerce: true, ATM: true},
		StepUpThreshold:   5_000_000,
	},
	{
		Code:               "CREDIT_GOLD",
		Name:               "SARD Credit Gold",
		Mode:               ModeCredit,
		BIN:                "735100",
		Currency:           "VND",
		ValidityYears:      4,
		TemporaryValidity:  24 * time.Hour,
		CardTypes:          []CardType{CardPlastic, CardVirtual, CardTemporary},
		MaxActiveVirtual:   3,
		DefaultCreditLimit: 50_000_000,
		MaxCreditLimit:     200_000_000,
		DefaultLimits:      Limits{PerTransaction: 50_000_000, Daily: 100_000_000},
		MaxLimits:          Limits{PerTransaction: 200_000_000, Daily: 300_000_000},
		Supported:          Controls{POS: true, Contactless: true, Ecommerce: true, ATM: true},
		StepUpThreshold:    1_000_000,
	},
	{
		Code:              "PREPAID_ONLINE",
		Name:              "SARD Prepaid Online",
		Mode:              ModePrepaid,
		BIN:               "739900",
		Currency:          "VND",
		ValidityYears:     3,
		TemporaryValidity: 24 * time.Hour,
		CardTypes:         []CardType{CardVirtual, CardTemporary},
		MaxActiveVirtual:  5,
		DefaultLimits:     Limits{PerTransaction: 5_000_000, Daily: 10_000_000},
		MaxLimits:         Limits{PerTransaction: 50_000_000, Daily: 100_000_000},
		Supported:         Controls{Ecommerce: true},
		StepUpThreshold:   0,
	},
}

type Catalog struct {
	byCode map[string]Program
	list   []Program
}

func NewCatalog(programs []Program) *Catalog {
	c := &Catalog{byCode: make(map[string]Program, len(programs))}
	for _, p := range programs {
		c.byCode[p.Code] = p
		c.list = append(c.list, p)
	}
	sort.Slice(c.list, func(i, j int) bool { return c.list[i].Code < c.list[j].Code })
	return c
}

func (c *Catalog) Get(code string) (Program, error) {
	p, ok := c.byCode[code]
	if !ok {
		return Program{}, Invalid("unknown program %q", code)
	}
	return p, nil
}

func (c *Catalog) All() []Program {
	return append([]Program(nil), c.list...)
}
