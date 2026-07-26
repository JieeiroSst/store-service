package model

type HealthStatus struct {
	MetadatabaseStatus       string `json:"metadatabase_status,omitempty"`
	SchedulerStatus          string `json:"scheduler_status,omitempty"`
	LatestSchedulerHeartbeat string `json:"latest_scheduler_heartbeat,omitempty"`
}
