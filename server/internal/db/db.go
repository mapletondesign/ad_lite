package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func Connect(dsn string) (*pgxpool.Pool, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return pool, nil
}

func ConnectRedis(url string) *redis.Client {
	opts, err := redis.ParseURL(url)
	if err != nil {
		// Fall back to default localhost if URL is missing or malformed.
		opts = &redis.Options{Addr: "localhost:6379"}
	}
	return redis.NewClient(opts)
}
