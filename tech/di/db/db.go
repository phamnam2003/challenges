package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewStorage(
	lc fx.Lifecycle,
	log *zap.Logger,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(
		"postgres://user:password@localhost:5432/appdb",
	)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 10
	cfg.MaxConnLifetime = time.Hour

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("postgres connecting")
			return pool.Ping(ctx)
		},
		OnStop: func(ctx context.Context) error {
			log.Info("postgres closing")
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
