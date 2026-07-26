package model

type TaskInstance struct {
	TaskId        string  `json:"task_id,omitempty"`
	DagId         string  `json:"dag_id,omitempty"`
	DagRunId      string  `json:"dag_run_id,omitempty"`
	ExecutionDate string  `json:"execution_date,omitempty"`
	StartDate     string  `json:"start_date,omitempty"`
	EndDate       string  `json:"end_date,omitempty"`
	Duration      float32 `json:"duration,omitempty"`
	State         string  `json:"state,omitempty"`
	TryNumber     int32   `json:"try_number,omitempty"`
	MapIndex      int32   `json:"map_index,omitempty"`
	Pool          string  `json:"pool,omitempty"`
	Operator      string  `json:"operator,omitempty"`
}

type TaskInstanceList struct {
	TaskInstances []TaskInstance `json:"task_instances"`
	TotalEntries  int32          `json:"total_entries"`
}
