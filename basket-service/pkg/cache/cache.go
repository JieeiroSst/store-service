package cache

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

type cacheHelper struct {
	resdis *redis.Client
}

type CacheHelper interface {
	GetInterface(ctx context.Context, key string, value interface{}) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, exppiration time.Duration) error
	Delete(ctx context.Context, key string) error
	HMSetComplex(ctx context.Context, key string, fields interface{}) error
	HMGetComplex(ctx context.Context, key string) ([]interface{}, error)
}

func NewCacheHelper(dns string) CacheHelper {
	rdb := redis.NewClient(&redis.Options{
		Addr:     dns,
		Password: "",
		DB:       0,
	})
	return &cacheHelper{
		resdis: rdb,
	}
}

func (h *cacheHelper) GetInterface(ctx context.Context, key string, value interface{}) (interface{}, error) {
	data, err := h.resdis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	typeValue := reflect.TypeOf(data)
	kind := typeValue.Kind()

	var outData interface{}
	switch kind {
	case reflect.Ptr, reflect.Struct, reflect.Slice:
		outData = reflect.New(typeValue).Interface()
	default:
		outData = reflect.Zero(typeValue).Interface()
	}
	err = json.Unmarshal([]byte(data), &outData)
	if err != nil {
		return nil, err
	}
	switch kind {
	case reflect.Ptr, reflect.Struct, reflect.Slice:
		return reflect.ValueOf(outData).Interface(), nil
	}
	var outValue interface{} = outData
	if reflect.TypeOf(outData).ConvertibleTo(typeValue) {
		outValueConverted := reflect.ValueOf(outData).Convert(typeValue)
		outValue = outValueConverted.Interface()
	}
	if outValue == nil {
		return nil, errors.New("")
	}
	return outData, nil
}

func (h *cacheHelper) Set(ctx context.Context, key string, value interface{}, exppiration time.Duration) error {
	data, err := json.Marshal(&value)
	if err != nil {
		return err
	}
	_ = h.resdis.Set(ctx, key, data, exppiration)
	return nil
}

func (h *cacheHelper) Delete(ctx context.Context, key string) error {
	value := h.resdis.Del(ctx, key)
	if value.Err() != nil {
		return value.Err()
	}
	return nil
}

func (h *cacheHelper) HMSetComplex(ctx context.Context, key string, fields interface{}) error {
	err := h.resdis.HMSet(ctx, key, fields).Err()
	if err != nil {
		return err
	}

	return nil
}

func (h *cacheHelper) HMGetComplex(ctx context.Context, key string) ([]interface{}, error) {
	values, err := h.resdis.HMGet(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	return values, nil
}
