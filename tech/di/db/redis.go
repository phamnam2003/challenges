package db

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewCache(life fx.Lifecycle) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	life.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return rdb.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			return rdb.Close()
		},
	})
	return rdb
}
