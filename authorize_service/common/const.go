package common

// RBACModelPath is the path to the Casbin RBAC model config.
const RBACModelPath = "config/conf/rbac_model.conf"

// Cache TTLs (in seconds).
const (
	CacheTTLEnforcer    = 300 // 5 minutes — enforcer rebuilt at most every 5 min
	CacheTTLCasbinByID  = 60  // 1 minute  — single-rule cache
)

// Cache key patterns.
const (
	CacheKeyCasbinByID = "casbin:rule:%d"
)
