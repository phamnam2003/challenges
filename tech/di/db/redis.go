package db

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewCache(
	lc fx.Lifecycle,
	log *zap.Logger,
) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("redis connecting")
			return rdb.Ping(ctx).Err()
		},
		OnStop: func(ctx context.Context) error {
			log.Info("redis closing")
			return rdb.Close()
		},
	})

	return rdb
}
