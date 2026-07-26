package model

import "time"

type DAGRun struct {
	DagId           string                 `json:"dag_id,omitempty"`
	DagRunId        string                 `json:"dag_run_id,omitempty"`
	LogicalDate     *time.Time             `json:"logical_date,omitempty"`
	StartDate       *time.Time             `json:"start_date,omitempty"`
	EndDate         *time.Time             `json:"end_date,omitempty"`
	State           string                 `json:"state,omitempty"`
	ExternalTrigger bool                   `json:"external_trigger,omitempty"`
	Conf            map[string]interface{} `json:"conf,omitempty"`
	Note            string                 `json:"note,omitempty"`
}

type DAGRunList struct {
	DagRuns      []DAGRun `json:"dag_runs"`
	TotalEntries int32    `json:"total_entries"`
}

type TriggerDAGRunRequest struct {
	DagRunId    string                 `json:"dag_run_id,omitempty"`
	LogicalDate *time.Time             `json:"logical_date,omitempty"`
	Conf        map[string]interface{} `json:"conf,omitempty"`
	Note        string                 `json:"note,omitempty"`
}
