package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	var pool *pgxpool.Pool
	for attempts := 1; attempts <= 10; attempts++ {
		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				return pool, nil
			}
			pool.Close()
		}
		if attempts < 10 {
			slog.Info("waiting for database...", "attempt", attempts)
			time.Sleep(time.Second)
		}
	}
	return nil, fmt.Errorf("ping db: %w", err)
}
