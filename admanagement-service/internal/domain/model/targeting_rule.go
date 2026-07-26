package model

import (
	"time"

	"gorm.io/datatypes"
)

type TargetingRuleType string

const (
	TargetingRuleTypeCountry TargetingRuleType = "country"
	TargetingRuleTypeAge     TargetingRuleType = "age"
	TargetingRuleTypeGender  TargetingRuleType = "gender"
	TargetingRuleTypeDevice  TargetingRuleType = "device"
	TargetingRuleTypeTime    TargetingRuleType = "time"
)

type TargetingRuleOperator string

const (
	TargetingOperatorEquals  TargetingRuleOperator = "equals"
	TargetingOperatorIn      TargetingRuleOperator = "in"
	TargetingOperatorBetween TargetingRuleOperator = "between"
	TargetingOperatorGreater TargetingRuleOperator = "greater"
	TargetingOperatorLess    TargetingRuleOperator = "less"
)

type AdTargetingRule struct {
	ID           uint                  `json:"id" gorm:"primaryKey"`
	AdID         uint                  `json:"ad_id"`
	RuleType     TargetingRuleType     `json:"rule_type"`
	RuleOperator TargetingRuleOperator `json:"rule_operator" gorm:"default:equals"`
	RuleValue    datatypes.JSON        `json:"rule_value"`
	IsActive     bool                  `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time             `json:"created_at"`
}

func (AdTargetingRule) TableName() string { return "ad_targeting_rules" }
