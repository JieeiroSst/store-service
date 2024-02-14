package router

import (
	"github.com/JIeeiroSst/oauth2-service/config"
	"github.com/JIeeiroSst/oauth2-service/internal/repository"
	"github.com/JIeeiroSst/oauth2-service/internal/usecase"
	"github.com/JIeeiroSst/oauth2-service/models"
	"github.com/JIeeiroSst/oauth2-service/pkg/token"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Router struct {
	config *config.Config
	db     *gorm.DB
	resdis *redis.Client
}

func NewRouter(config *config.Config, db *gorm.DB, resdis *redis.Client) *Router {
	return &Router{
		config: config,
		db:     db,
		resdis: resdis,
	}
}

func (r *Router) Init() {
	usecase := usecase.NewDefaultManager()
	usecase.SetAuthorizeCodeTokenCfg(config.DefaultAuthorizeCodeTokenCfg)

	usecase.MustTokenStorage(repository.NewMemoryTokenStore(r.db, r.resdis))

	usecase.MapAccessGenerate(token.NewAccessGenerate())

	clientStore := repository.NewClientStore()
	clientStore.Set(r.config.Secret.Idvar, &models.Client{
		ID:     r.config.Secret.Idvar,
		Secret: r.config.Secret.Secretvar,
		Domain: r.config.Secret.Domainvar,
	})
}
