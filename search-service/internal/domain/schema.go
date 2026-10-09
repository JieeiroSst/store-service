package domain

type SchemaStatus struct {
	Template string      `json:"template"`
	Applied  bool        `json:"applied"`
	Version  int         `json:"version"`
	Indices  []IndexInfo `json:"indices"`
}

type ReindexResult struct {
	Name   string `json:"name"`
	From   string `json:"from"`
	To     string `json:"to"`
	Copied int64  `json:"copied"`
	Error  string `json:"error,omitempty"`
}
