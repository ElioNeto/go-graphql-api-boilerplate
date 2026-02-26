.PHONY: build run test tidy generate migrate-up migrate-down docker-build docker-run lint

# Variables
APP_NAME    := go-graphql-api
BIN_DIR     := ./bin
MAIN_PATH   := ./cmd/api
DOCKER_IMG  := go-graphql-api-boilerplate
DB_DSN      ?= $(shell grep DB_HOST .env 2>/dev/null | cut -d '=' -f2)

build: generate
	@echo ">> Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) $(MAIN_PATH)

run: generate
	@echo ">> Running $(APP_NAME)..."
	go run $(MAIN_PATH)

generate:
	@echo ">> Generating GraphQL code..."
	go run github.com/99designs/gqlgen generate

test:
	@echo ">> Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

tidy:
	@echo ">> Tidying modules..."
	go mod tidy

lint:
	@echo ">> Running linter..."
	golangci-lint run ./...

migrate-up:
	@echo ">> Running migrations up..."
	golang-migrate -path ./migrations -database "$$(cat .env | grep -E '^DB_' | xargs)" up

migrate-down:
	@echo ">> Running migrations down..."
	golang-migrate -path ./migrations -database "$$(cat .env | grep -E '^DB_' | xargs)" down 1

docker-build:
	@echo ">> Building Docker image..."
	docker build -t $(DOCKER_IMG):latest .

docker-run:
	@echo ">> Running Docker container..."
	docker run --env-file .env -p 8080:8080 $(DOCKER_IMG):latest
