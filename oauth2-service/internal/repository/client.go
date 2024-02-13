package repository

import (
	"sync"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type ClientRepository struct {
	sync.RWMutex
	db     *gorm.DB
	resdis *redis.Client
}

func NewClientRepository(db *gorm.DB,
	resdis *redis.Client) *ClientRepository {
	return &ClientRepository{
		db:     db,
		resdis: resdis,
	}
}
