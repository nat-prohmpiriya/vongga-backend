package adapter

import (
	"context"
	"vongga-api/config"

	"github.com/redis/go-redis/v9"
)

func InitRedis(config *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisURI,
		Password: config.RedisPassword,
		DB:       0,
	})

	// Test connection
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}
