package model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestUserCouponBelongsToCoupon(t *testing.T) {
	s, err := schema.Parse(&UserCoupon{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	rel, ok := s.Relationships.Relations["Coupon"]
	if !ok {
		t.Fatal("Coupon relation missing")
	}
	if rel.Type != schema.BelongsTo {
		t.Fatalf("relation type = %s, want belongs_to", rel.Type)
	}
	if got := rel.References[0].ForeignKey.DBName; got != "coupon_id" {
		t.Fatalf("foreign key = %s, want coupon_id", got)
	}
}
