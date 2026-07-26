package di

import (
	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/jwttoken"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/pg"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/pwhash"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/sessionstore"
	"github.com/JIeeiroSst/user-service/internal/application/auth"
	"github.com/JIeeiroSst/user-service/internal/application/role"
	"github.com/JIeeiroSst/user-service/internal/application/roleitem"
	"github.com/JIeeiroSst/user-service/internal/application/user"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/cache/expire"
	pgconn "github.com/JIeeiroSst/utils/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const ecosystem = ".env"

func loadConfig() (*config.Config, error) {
	return config.InitializeConfiguration(ecosystem)
}

func newDB(cfg *config.Config) (*gorm.DB, error) {
	db := pgconn.NewPostgresConn(pgconn.PostgresConfig{
		PostgresqlHost:     cfg.Postgres.PostgresqlHost,
		PostgresqlPort:     cfg.Postgres.PostgresqlPort,
		PostgresqlUser:     cfg.Postgres.PostgresqlUser,
		PostgresqlPassword: cfg.Postgres.PostgresqlPassword,
		PostgresqlDbname:   cfg.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  cfg.Postgres.PostgresqlSSLMode,
	})
	if err := db.AutoMigrate(&domain.User{}, &domain.Role{}); err != nil {
		return nil, err
	}
	return db, nil
}

func newCache(cfg *config.Config) expire.CacheHelper {
	return expire.NewCacheHelper(cfg.Redis.Dns)
}

func newUserRepository(db *gorm.DB) output.UserRepository { return pg.NewUserRepository(db) }

func newRoleRepository(db *gorm.DB) output.RoleRepository { return pg.NewRoleRepository(db) }

func newRoleItemRepository(db *gorm.DB) output.RoleItemRepository {
	return pg.NewRoleItemRepository(db)
}

func newHasher() output.Hasher { return pwhash.New() }

func newTokenGenerator(cfg *config.Config) output.TokenGenerator {
	return jwttoken.New(cfg.Secret.JwtSecretKey, cfg.Token.AccessTokenTTL())
}

func newTokenStore(cache expire.CacheHelper) output.TokenStore {
	return sessionstore.New(cache)
}

func newAuthService(cfg *config.Config, userRepo output.UserRepository, hasher output.Hasher, tokenGen output.TokenGenerator, tokenStore output.TokenStore) input.AuthService {
	return auth.New(userRepo, hasher, tokenGen, tokenStore, cfg.Token.AccessTokenTTL(), cfg.Token.RefreshTokenTTL())
}

func newUserService(userRepo output.UserRepository, hasher output.Hasher, cache expire.CacheHelper) input.UserService {
	return user.New(userRepo, hasher, cache)
}

func newRoleService(roleRepo output.RoleRepository) input.RoleService { return role.New(roleRepo) }

func newRoleItemService(roleItemRepo output.RoleItemRepository) input.RoleItemService {
	return roleitem.New(roleItemRepo)
}

func newGRPCHandler(authSvc input.AuthService, userSvc input.UserService, roleSvc input.RoleService, roleItemSvc input.RoleItemService) *grpcadapter.Handler {
	return grpcadapter.NewHandler(authSvc, userSvc, roleSvc, roleItemSvc)
}

var Module = fx.Options(
	fx.Provide(
		loadConfig,
		newDB,
		newCache,
		newUserRepository,
		newRoleRepository,
		newRoleItemRepository,
		newHasher,
		newTokenGenerator,
		newTokenStore,
		newAuthService,
		newUserService,
		newRoleService,
		newRoleItemService,
		newGRPCHandler,
	),
	fx.Invoke(RegisterServer),
)
