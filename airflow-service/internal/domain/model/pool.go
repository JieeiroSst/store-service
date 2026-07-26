package model

type Pool struct {
	Name          string `json:"name,omitempty"`
	Slots         int32  `json:"slots"`
	OccupiedSlots int32  `json:"occupied_slots,omitempty"`
	UsedSlots     int32  `json:"used_slots,omitempty"`
	QueuedSlots   int32  `json:"queued_slots,omitempty"`
	OpenSlots     int32  `json:"open_slots,omitempty"`
	Description   string `json:"description,omitempty"`
}

type PoolList struct {
	Pools        []Pool `json:"pools"`
	TotalEntries int32  `json:"total_entries"`
}
