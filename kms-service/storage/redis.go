package storage

import (
	"context"
	"encoding/json"
	"time"
	
	"github.com/JIeeiroSst/kms/config"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	SetKey(string, []byte, time.Duration) error
	GetKey(string) ([]byte, error)
	DeleteKey(string) error
	Set(string, interface{}, time.Duration) error
	Get(string) (string, error)
	Close() error
}

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache() (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.RedisAddr,
		Password: config.AppConfig.RedisPassword,
		DB:       0,
	})
	
	ctx := context.Background()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	
	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisCache) SetKey(keyID string, encryptedKey []byte, expiration time.Duration) error {
	return r.client.Set(r.ctx, "key:"+keyID, encryptedKey, expiration).Err()
}

func (r *RedisCache) GetKey(keyID string) ([]byte, error) {
	val, err := r.client.Get(r.ctx, "key:"+keyID).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (r *RedisCache) DeleteKey(keyID string) error {
	return r.client.Del(r.ctx, "key:"+keyID).Err()
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, expiration).Err()
}

func (r *RedisCache) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}