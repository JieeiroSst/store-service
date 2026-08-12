package model

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type ConstraintOperator string

const (
	OperatorIn          ConstraintOperator = "IN"
	OperatorNotIn       ConstraintOperator = "NOT_IN"
	OperatorStrContains ConstraintOperator = "STR_CONTAINS"
)

type Constraint struct {
	ID              uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StrategyID      uuid.UUID          `gorm:"type:uuid;not null;index" json:"strategyId"`
	ContextField    string             `gorm:"not null" json:"contextField"`
	Operator        ConstraintOperator `gorm:"not null" json:"operator"`
	Values          datatypes.JSON     `gorm:"type:jsonb" json:"values"`
	CaseInsensitive bool               `json:"caseInsensitive"`
}

func (Constraint) TableName() string { return "constraints" }
