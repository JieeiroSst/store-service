package config

import "time"

type ConfigToken struct {
	AccessTokenExp    time.Duration
	RefreshTokenExp   time.Duration
	IsGenerateRefresh bool
}

type RefreshingConfig struct {
	AccessTokenExp     time.Duration
	RefreshTokenExp    time.Duration
	IsGenerateRefresh  bool
	IsResetRefreshTime bool
	IsRemoveAccess     bool
	IsRemoveRefreshing bool
}

var (
	DefaultCodeExp               = time.Minute * 10
	DefaultAuthorizeCodeTokenCfg = &ConfigToken{AccessTokenExp: time.Hour * 2, RefreshTokenExp: time.Hour * 24 * 3, IsGenerateRefresh: true}
	DefaultImplicitTokenCfg      = &ConfigToken{AccessTokenExp: time.Hour * 1}
	DefaultPasswordTokenCfg      = &ConfigToken{AccessTokenExp: time.Hour * 2, RefreshTokenExp: time.Hour * 24 * 7, IsGenerateRefresh: true}
	DefaultClientTokenCfg        = &ConfigToken{AccessTokenExp: time.Hour * 2}
	DefaultRefreshTokenCfg       = &RefreshingConfig{IsGenerateRefresh: true, IsRemoveAccess: true, IsRemoveRefreshing: true}
)
