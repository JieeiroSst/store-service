package model

type DAG struct {
	DagId       string   `json:"dag_id,omitempty"`
	IsPaused    bool     `json:"is_paused"`
	IsActive    bool     `json:"is_active,omitempty"`
	IsSubdag    bool     `json:"is_subdag,omitempty"`
	Owners      []string `json:"owners,omitempty"`
	Description string   `json:"description,omitempty"`
	Fileloc     string   `json:"fileloc,omitempty"`
}

type DAGList struct {
	Dags         []DAG `json:"dags"`
	TotalEntries int32 `json:"total_entries"`
}
