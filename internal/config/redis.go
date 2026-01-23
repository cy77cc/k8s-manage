package config

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

func MustNewRedisClient() redis.UniversalClient {
	if !CFG.Redis.Enable {
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         CFG.Redis.Addr,
		Password:     CFG.Redis.Password,
		DB:           CFG.Redis.DB,
		PoolSize:     CFG.Redis.PoolSize,
		MinIdleConns: CFG.Redis.MinIdleConns,
		DialTimeout:  CFG.Redis.DialTimeout,
		ReadTimeout:  CFG.Redis.ReadTimeout,
		WriteTimeout: CFG.Redis.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect redis: %v", err)
	}

	return rdb
}
