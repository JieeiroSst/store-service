package model

// V0 = subject/role, V1 = object/endpoint, V2 = action/HTTP method.
type CasbinRule struct {
	ID    int    `gorm:"primaryKey;autoIncrement" json:"id"`
	Ptype string `json:"ptype"`
	V0    string `json:"v0"` // subject (role/user)
	V1    string `json:"v1"` // object (endpoint path)
	V2    string `json:"v2"` // action (HTTP method)
	V3    string `json:"v3,omitempty"`
	V4    string `json:"v4,omitempty"`
	V5    string `json:"v5,omitempty"`
}

func (CasbinRule) TableName() string { return "casbin_rule" }

// CasbinAuth holds the (subject, object, action) triple used for enforcement.
type CasbinAuth struct {
	Sub string `json:"sub"` // who  (username / role)
	Obj string `json:"obj"` // what (endpoint path)
	Act string `json:"act"` // how  (GET / POST / …)
}

// UpdateField is a typed enum for columns that may be updated individually.
type UpdateField string

const (
	FieldPtype    UpdateField = "ptype"
	FieldSubject  UpdateField = "v0"
	FieldEndpoint UpdateField = "v1"
	FieldMethod   UpdateField = "v2"
)

func (f UpdateField) IsValid() bool {
	switch f {
	case FieldPtype, FieldSubject, FieldEndpoint, FieldMethod:
		return true
	}
	return false
}
