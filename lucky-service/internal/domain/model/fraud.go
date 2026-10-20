package model

import (
	"strings"
	"time"
)

type ClientInfo struct {
	IP        string `json:"ip"`
	DeviceID  string `json:"device_id"`
	UserAgent string `json:"user_agent"`
}

func (c ClientInfo) Normalize() ClientInfo {
	c.IP = clip(strings.TrimSpace(c.IP), 64)
	c.DeviceID = clip(strings.TrimSpace(c.DeviceID), 128)
	c.UserAgent = clip(strings.TrimSpace(c.UserAgent), 255)
	return c
}

func clip(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

type RewardStatus string

const (
	RewardPending  RewardStatus = "pending"
	RewardHeld     RewardStatus = "held"
	RewardRejected RewardStatus = "rejected"
)

func (s RewardStatus) Valid() bool {
	return s == RewardPending || s == RewardHeld || s == RewardRejected
}

const (
	HoldLargeReward = "large_reward"
	HoldSharedIP    = "shared_ip"
	HoldBlockedUser = "blocked_user"
)

type UserBlock struct {
	UserID    string    `json:"user_id"`
	Reason    string    `json:"reason"`
	BlockedBy string    `json:"blocked_by"`
	CreatedAt time.Time `json:"created_at"`
}

type SignalKind string

const (
	SignalDevice SignalKind = "device"
	SignalIP     SignalKind = "ip"
)

type FraudPolicy struct {
	RequireDevice     bool
	MaxUsersPerDevice int
	MaxUsersPerIP     int
}
