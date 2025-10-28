package pkg

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	client *redis.Client
	prefix string
}

func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:	addr,
		Password: password,
		DB:	db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisClient{
		client: rdb,
		prefix: "lab1:",
	}, nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) AddToBlacklist(ctx context.Context, token string, expiration time.Duration) error {
    if r.client == nil {
        return fmt.Errorf("redis client is not initialized")
    }
    
    key := r.prefix + "blacklist:" + token
    logrus.Infof("Setting Redis key: %s with TTL: %v", key, expiration)
    
    err := r.client.Set(ctx, key, "1", expiration).Err()
    if err != nil {
        logrus.Errorf("Failed to set Redis key %s: %v", key, err)
        return err
    }
    
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        logrus.Errorf("Failed to verify Redis key %s: %v", key, err)
        return err
    }
    
    logrus.Infof("Successfully set Redis key: %s, value: %s", key, val)
    return nil
}

func (r *RedisClient) IsInBlacklist(ctx context.Context, token string) (bool, error) {
	key := r.prefix + "blacklist:" + token
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return val == "1", nil
}