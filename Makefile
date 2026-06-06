.PHONY: build run swag test tidy vet ensure-tools migrate-new migrate-up migrate-down migrate-fresh

# Auto-install required CLI tools
ensure-tools:
	@which swag > /dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@latest
	@which migrate > /dev/null 2>&1 || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Build the binary
build:
	go build -o bin/server ./cmd/http

# Run the server
run:
	go run ./cmd/http

# Generate Swagger docs (auto-installs swag if missing)
swag: ensure-tools
	swag init -g cmd/http/main.go -o ./docs --parseDependency --parseInternal

# Run all tests
test:
	go test ./...

# Tidy module dependencies
tidy:
	go mod tidy

# Run static analysis
vet:
	go vet ./...

# Create a new migration: make migrate-new name=create_users_table
migrate-new:
	@which migrate > /dev/null 2>&1 || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	migrate create -ext sql -dir files/migrations -seq $(name)

# Run all pending migrations
migrate-up:
	go run ./cmd/migrate up

# Roll back the last migration
migrate-down:
	go run ./cmd/migrate down

# Drop all tables then re-run all migrations
migrate-fresh:
	go run ./cmd/migrate fresh