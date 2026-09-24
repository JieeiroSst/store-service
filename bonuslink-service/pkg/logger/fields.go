package logger

import "go.uber.org/zap"

func RefCode(v string) zap.Field { return zap.String("ref_code", v) }

func OwnerUserID(v string) zap.Field { return zap.String("owner_user_id", v) }

func NewUserID(v string) zap.Field { return zap.String("new_user_id", v) }

func EventType(v string) zap.Field { return zap.String("event_type", v) }

func Channel(v string) zap.Field { return zap.String("channel", v) }

func Platform(v string) zap.Field { return zap.String("platform", v) }

func DeepLink(v string) zap.Field { return zap.String("deep_link", v) }

func RewardValue(v float64) zap.Field { return zap.Float64("reward_value", v) }

func RewardType(v string) zap.Field { return zap.String("reward_type", v) }

func RequestID(v string) zap.Field { return zap.String("request_id", v) }

func StatusCode(v int) zap.Field { return zap.Int("status_code", v) }

func Table(v string) zap.Field { return zap.String("table", v) }
