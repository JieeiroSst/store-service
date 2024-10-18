package dto

import "github.com/JieeiroSst/authorize-service/model"

type CasbinRule struct {
	ID    int `gorm:"primaryKey;autoIncrement"`
	Ptype string
	V0    string
	V1    string
	V2    string
	V3    string
	V4    string
	V5    string
}

type CasbinAuth struct {
	Sub string
	Obj string
	Act string
}

func (c CasbinRule) Build() model.CasbinRule {
	return model.CasbinRule{
		ID:    c.ID,
		Ptype: c.Ptype,
		V0:    c.V0,
		V1:    c.V1,
		V2:    c.V2,
		V3:    c.V3,
		V4:    c.V4,
		V5:    c.V5,
	}
}

func (c CasbinAuth) Build() model.CasbinAuth {
	return model.CasbinAuth{
		Sub: c.Sub,
		Obj: c.Obj,
		Act: c.Act,
	}
}
