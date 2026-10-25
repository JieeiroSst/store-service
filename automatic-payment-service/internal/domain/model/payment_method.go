package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PaymentMethod struct {
	PaymentMethodID uuid.UUID `gorm:"column:payment_method_id;primaryKey" json:"paymentMethodId"`
	UserID          uuid.UUID `gorm:"column:user_id;index;not null" json:"userId"`
	Provider        string    `gorm:"column:provider;not null" json:"provider"` // visa, mastercard, paypal, ...
	TokenID         string    `gorm:"column:token_id" json:"tokenId,omitempty"`
	LastFourDigits  string    `gorm:"column:last_four_digits" json:"lastFourDigits,omitempty"`
	ExpiryDate      string    `gorm:"column:expiry_date" json:"expiryDate,omitempty"` // MM/YYYY
	IsDefault       bool      `gorm:"column:is_default;not null;default:false" json:"isDefault"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (PaymentMethod) TableName() string { return "payment_methods" }

var (
	lastFourPattern   = regexp.MustCompile(`^[0-9]{4}$`)
	expiryDatePattern = regexp.MustCompile(`^(0[1-9]|1[0-2])/[0-9]{4}$`)
)

func ValidLastFour(s string) bool { return s == "" || lastFourPattern.MatchString(s) }

func ValidExpiryDate(s string) bool { return s == "" || expiryDatePattern.MatchString(s) }

func LooksLikePAN(s string) bool {
	digits := strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(s))
	if len(digits) < 13 || len(digits) > 19 {
		return false
	}
	sum := 0
	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if d < 0 || d > 9 {
			return false
		}
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return sum%10 == 0
}
