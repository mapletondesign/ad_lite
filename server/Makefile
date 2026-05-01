.PHONY: dev up down build test lint

# Start local Postgres + Redis
up:
	docker compose up -d

# Stop local services
down:
	docker compose down

# Run the API server (requires services to be up)
dev:
	go run ./cmd/api/...

# Build binary
build:
	go build -o bin/api ./cmd/api/...

# Run tests
test:
	go test ./...

# Lint (requires golangci-lint: brew install golangci-lint)
lint:
	golangci-lint run ./...

# Wipe and recreate the database (destructive — dev only)
reset-db:
	docker compose down -v
	docker compose up -d
