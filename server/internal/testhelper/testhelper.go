package testhelper

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// DB returns a pool connected to the test database with migrations applied.
// Uses TEST_DATABASE_URL env var; falls back to a local default.
func DB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://adlite:adlite@localhost:5432/ad_lite_test?sslmode=disable"
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect test db: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping test db (is docker compose up?): %v", err)
	}
	runMigrations(t, pool)
	t.Cleanup(func() { pool.Close() })
	return pool
}

// Redis returns a client connected to the test Redis database (DB index 1).
// Uses TEST_REDIS_URL env var; falls back to a local default.
func Redis(t *testing.T) *redis.Client {
	t.Helper()
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		url = "redis://localhost:6379/1"
	}
	opts, err := redis.ParseURL(url)
	if err != nil {
		t.Fatalf("parse test redis URL: %v", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping test redis (is docker compose up?): %v", err)
	}
	t.Cleanup(func() {
		rdb.FlushDB(context.Background())
		rdb.Close()
	})
	return rdb
}

// RSAKeys generates a fresh RSA key pair for use in auth/JWT tests.
func RSAKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	return priv, &priv.PublicKey
}

// TruncateAll removes all rows from every table, resetting the test DB to empty.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		TRUNCATE admin_audit_log, impressions, refresh_tokens, users, bookings, ad_slots, advertisers, devices, venues CASCADE
	`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

// runMigrations applies any unapplied SQL migration files in lexicographic order.
// Uses a schema_migrations table to track what has already been applied, so
// new migrations added after the test DB was first created are picked up
// automatically on the next test run.
func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename   TEXT        PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`,
	); err != nil {
		t.Fatalf("create schema_migrations: %v", err)
	}

	_, file, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir %s: %v", migrationsDir, err)
	}

	var sqlFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, e.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, name := range sqlFiles {
		var applied bool
		_ = pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`, name,
		).Scan(&applied)
		if applied {
			continue
		}
		sql, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			// "already exists" means the migration ran before the tracker was
			// introduced — treat it as applied rather than failing.
			if !strings.Contains(err.Error(), "already exists") {
				t.Fatalf("apply migration %s: %v", name, err)
			}
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`, name,
		); err != nil {
			t.Fatalf("record migration %s: %v", name, err)
		}
	}
}
