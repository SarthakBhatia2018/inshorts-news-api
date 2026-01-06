.PHONY: build run migrate-up migrate-down migrate-down-all migrate-reset migrate-status migrate-create migrate-version load-data test clean docker-up docker-down setup

# Build binaries
build:
	go build -o bin/server main.go
	go build -o bin/migrate cmd/migrate/main.go

# Run the server
run:
	go run main.go

# Migration commands

# Run all pending migrations
migrate-up:
	@echo "Running all pending migrations..."
	go run cmd/migrate/main.go up

# Rollback the last migration (remove one)
migrate-down:
	@echo "Rolling back last migration..."
	go run cmd/migrate/main.go down

# Rollback all migrations
migrate-down-all:
	@echo "Rolling back all migrations..."
	go run cmd/migrate/main.go reset

# Alternative name for reset (removes all migrations)
migrate-reset: migrate-down-all

# Run only the next pending migration (up by one)
migrate-up-one:
	@echo "Running next migration..."
	go run cmd/migrate/main.go up-by-one

# Show migration status
migrate-status:
	@echo "Checking migration status..."
	go run cmd/migrate/main.go status

# Show current migration version
migrate-version:
	@echo "Current migration version:"
	go run cmd/migrate/main.go version

# Load news data from JSON file
load-data:
	@echo "Loading news data..."
	go run scripts/load_data.go
