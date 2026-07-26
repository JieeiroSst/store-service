package model

type Variable struct {
	Key         string `json:"key,omitempty"`
	Value       string `json:"value,omitempty"`
	Description string `json:"description,omitempty"`
}

type VariableList struct {
	Variables    []Variable `json:"variables"`
	TotalEntries int32      `json:"total_entries"`
}
