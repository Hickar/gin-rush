package cache

import (
	"context"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/go-redis/redis/v8"
)

var _redisClient *redis.Client

func NewCache(ctx context.Context, conf *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:               ":6379",
		Password:           conf.Password,
		DB:                 conf.Db,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	_redisClient = client
	return _redisClient, nil
}

func GetCache() *redis.Client {
	return _redisClient
}