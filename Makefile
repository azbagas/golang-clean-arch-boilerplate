.PHONY: run build test lint swagger migrate-up migrate-down docker-up docker-down clean

# Application
APP_NAME=server
CMD_PATH=./cmd/api

# Run the application
run:
	go run $(CMD_PATH)/main.go

# Build the application
build:
	go build -ldflags="-w -s" -o bin/$(APP_NAME) $(CMD_PATH)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	go vet ./...

# Generate swagger docs (requires swag CLI: go install github.com/swaggo/swag/cmd/swag@latest)
swagger:
	swag init -g cmd/api/main.go -o docs

# Database migrations (requires golang-migrate CLI)
migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/clean_arch_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/clean_arch_db?sslmode=disable" down

# Docker
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ coverage.out coverage.html
