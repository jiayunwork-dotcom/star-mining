package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisCache(addr string, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

func (r *RedisCache) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	switch v := value.(type) {
	case string:
		return r.client.Set(r.ctx, key, v, expiration).Err()
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		return r.client.Set(r.ctx, key, data, expiration).Err()
	}
}

func (r *RedisCache) Del(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

func (r *RedisCache) Exists(key string) (bool, error) {
	n, err := r.client.Exists(r.ctx, key).Result()
	return n > 0, err
}

func (r *RedisCache) GetJSON(key string, dest interface{}) error {
	data, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

func (r *RedisCache) SetJSON(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(r.ctx, key, data, expiration).Err()
}

func (r *RedisCache) HSet(key string, field string, value interface{}) error {
	return r.client.HSet(r.ctx, key, field, value).Err()
}

func (r *RedisCache) HGet(key string, field string) (string, error) {
	return r.client.HGet(r.ctx, key, field).Result()
}

func (r *RedisCache) HGetAll(key string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, key).Result()
}

func (r *RedisCache) HDel(key string, fields ...string) error {
	return r.client.HDel(r.ctx, key, fields...).Err()
}

func (r *RedisCache) SAdd(key string, members ...interface{}) error {
	return r.client.SAdd(r.ctx, key, members...).Err()
}

func (r *RedisCache) SRem(key string, members ...interface{}) error {
	return r.client.SRem(r.ctx, key, members...).Err()
}

func (r *RedisCache) SMembers(key string) ([]string, error) {
	return r.client.SMembers(r.ctx, key).Result()
}

func (r *RedisCache) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

func (r *RedisCache) GetClient() *redis.Client {
	return r.client
}

func (r *RedisCache) GetContext() context.Context {
	return r.ctx
}
