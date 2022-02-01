package cache

import (
	"context"

	"github.com/Hickar/gin-rush/internal/config"
	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redismock/v8"
)

func NewCache(ctx context.Context, conf *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:               conf.Host,
		Password:           conf.Password,
		DB:                 conf.Db,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func NewCacheMock() (*redis.Client, redismock.ClientMock) {
	return redismock.NewClientMock()
}