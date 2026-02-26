# Getting Started with Go GraphQL API Boilerplate

Complete guide to set up and run the GraphQL API locally and in development environments.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Database Setup](#database-setup)
- [GraphQL Code Generation](#graphql-code-generation)
- [Running the Application](#running-the-application)
- [Accessing GraphQL Playground](#accessing-graphql-playground)
- [Understanding the Architecture](#understanding-the-architecture)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- **Go 1.22+**: [Download](https://go.dev/dl/)
- **PostgreSQL 15+**: [Download](https://www.postgresql.org/download/)
- **Make**: Usually pre-installed on Unix systems
- **Git**: [Download](https://git-scm.com/downloads)

### Verify Installation

```bash
# Check Go version
go version
# Expected: go version go1.22.x

# Check PostgreSQL
psql --version
# Expected: psql (PostgreSQL) 15.x

# Check Make
make --version
# Expected: GNU Make 4.x
```

## Installation

### Step 1: Clone the Repository

```bash
git clone https://github.com/ElioNeto/go-graphql-api-boilerplate.git
cd go-graphql-api-boilerplate
```

### Step 2: Install Go Dependencies

```bash
# Download all dependencies
go mod download

# Verify dependencies
go mod verify

# Or use Make
make tidy
```

### Step 3: Configure Environment Variables

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Application
APP_HOST=0.0.0.0
APP_PORT=8080
APP_ENV=development
APP_DEBUG=true

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=graphql_db
DB_SSLMODE=disable
DB_MAX_CONNECTIONS=25
DB_MIGRATIONS_PATH=file://migrations

# Authentication
AUTH_JWT_SECRET=your_secret_key_change_in_production
AUTH_JWT_EXPIRATION=24

# GraphQL
GRAPHQL_PLAYGROUND_ENABLED=true
GRAPHQL_INTROSPECTION_ENABLED=true
```

### Step 4: Generate JWT Secret

For security, generate a strong JWT secret:

```bash
# Using openssl
openssl rand -hex 32

# Or using Go
go run -c 'package main; import ("crypto/rand"; "encoding/hex"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(hex.EncodeToString(b)) }'

# Add to .env
echo "AUTH_JWT_SECRET=$(openssl rand -hex 32)" >> .env
```

## Database Setup

### Option 1: Using PostgreSQL CLI

```bash
# Create database
createdb -U postgres graphql_db

# Verify creation
psql -U postgres -l | grep graphql_db
```

### Option 2: Using psql

```bash
# Connect to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE graphql_db;

# Connect to the database
\c graphql_db

# Verify
\l

# Exit
\q
```

### Option 3: Using Docker

```bash
# Run PostgreSQL in Docker
docker run -d \
  --name postgres-graphql \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=graphql_db \
  -p 5432:5432 \
  postgres:15-alpine

# Check logs
docker logs postgres-graphql
```

### Run Migrations

Apply database migrations:

```bash
# Using Make (recommended)
make migrate-up

# Or manually install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations manually
migrate -path migrations \
  -database "postgresql://postgres:postgres@localhost:5432/graphql_db?sslmode=disable" \
  up
```

### Verify Tables

```bash
psql -U postgres graphql_db -c "\dt"

# Expected output:
#              List of relations
#  Schema |       Name        | Type  |  Owner   
# --------+-------------------+-------+----------
#  public | schema_migrations | table | postgres
#  public | users             | table | postgres
```

## GraphQL Code Generation

**Important**: This project uses [gqlgen](https://gqlgen.com/) for code generation. You must regenerate GraphQL code whenever you modify the schema.

### Generate GraphQL Code

```bash
# Using Make (recommended)
make generate

# Or directly
go run github.com/99designs/gqlgen generate
```

### What Gets Generated?

- `graph/generated.go` - Core GraphQL execution code
- `graph/model/models_gen.go` - GraphQL input/output types
- `graph/schema.resolvers.go` - Resolver stubs (only if missing)

### When to Regenerate?

- After modifying `graph/schema.graphqls`
- After changing the configuration in `gqlgen.yml`
- When adding new types, queries, or mutations

## Running the Application

### Development Mode

```bash
# Using Make
make run

# Or with Go directly
go run cmd/api/main.go

# With live reload using air
air
```

Expected output:

```
2026/02/26 18:00:00 INFO Starting GraphQL API server
2026/02/26 18:00:00 INFO Environment: development
2026/02/26 18:00:00 INFO Database connected host=localhost
2026/02/26 18:00:00 INFO Migrations applied successfully
2026/02/26 18:00:00 INFO GraphQL Playground enabled at http://localhost:8080/
2026/02/26 18:00:00 INFO Server listening on :8080
```

### Build for Production

```bash
# Build binary
make build

# Run the binary
./bin/api
```

## Accessing GraphQL Playground

The GraphQL Playground is an interactive IDE for exploring the API.

### Open Playground

1. Start the server: `make run`
2. Open browser: [http://localhost:8080/](http://localhost:8080/)
3. You should see the GraphQL Playground interface

### First Query - Health Check

Try this simple query:

```graphql
query {
  health
}
```

Expected response:

```json
{
  "data": {
    "health": "ok"
  }
}
```

### Explore the Schema

Click the "Schema" button on the right to see:
- Available queries
- Available mutations
- Type definitions
- Input types

### Using Query Variables

Variables make queries reusable:

```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    name
    email
  }
}
```

Query Variables panel:

```json
{
  "input": {
    "name": "Alice",
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }
}
```

## Understanding the Architecture

### Request Flow

```
Client Request
    ↓
HTTP Middleware (Logging, Recovery)
    ↓
Dataloader Middleware (N+1 prevention)
    ↓
JWT Auth Middleware (if protected)
    ↓
GraphQL Handler (gqlgen)
    ↓
Resolver (graph/schema.resolvers.go)
    ↓
Service Layer (internal/services)
    ↓
Repository Layer (internal/repositories)
    ↓
PostgreSQL Database
```

### Layer Responsibilities

| Layer | Responsibility | Example |
|-------|----------------|----------|
| **Resolver** | GraphQL request/response mapping | Parse input, call service, format output |
| **Service** | Business logic | Validate data, orchestrate operations |
| **Repository** | Data access | SQL queries, database interactions |
| **Dataloader** | Batch & cache requests | Prevent N+1 queries |

### Key Files

- `graph/schema.graphqls` - GraphQL schema definition
- `graph/schema.resolvers.go` - Resolver implementations
- `internal/services/user_service.go` - Business logic
- `internal/repositories/user_repository.go` - Database operations
- `internal/dataloaders/loaders.go` - N+1 prevention
- `cmd/api/main.go` - Application entry point

## Troubleshooting

### Issue: "database connection refused"

**Cause**: PostgreSQL is not running

**Solution**:

```bash
# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Start PostgreSQL (macOS)
brew services start postgresql@15

# Start PostgreSQL (Linux)
sudo systemctl start postgresql

# Start PostgreSQL (Docker)
docker start postgres-graphql
```

### Issue: "migration: no change"

**Cause**: Migrations already applied

**Solution**: This is normal. If you need to reapply:

```bash
make migrate-down
make migrate-up
```

### Issue: "gqlgen: command not found"

**Cause**: gqlgen not installed

**Solution**:

```bash
go install github.com/99designs/gqlgen@latest

# Verify
gqlgen version
```

### Issue: "validation failed: field ... not defined"

**Cause**: Schema and resolvers out of sync

**Solution**:

```bash
# Regenerate GraphQL code
make generate

# If that fails, remove generated files first
rm graph/generated.go graph/model/models_gen.go
make generate
```

### Issue: Port 8080 already in use

**Cause**: Another process using port 8080

**Solution**:

```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or change port in .env
echo "APP_PORT=8081" >> .env
```

### Issue: "invalid JWT token"

**Cause**: JWT secret mismatch or expired token

**Solution**:

1. Verify `AUTH_JWT_SECRET` in `.env`
2. Login again to get a fresh token
3. Check token expiration in `.env` (default: 24 hours)

### Issue: Go module errors

**Solution**:

```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy up
go mod tidy
```

### Enable Debug Logging

For more detailed logs:

```env
APP_DEBUG=true
```

This will show:
- SQL queries
- Dataloader batching
- Request/response details

## Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/services/... -v

# Run with race detector
go test -race ./...
```

## Next Steps

Now that your API is running:

1. **Explore GraphQL**: Check [GraphQL Examples](./GRAPHQL_EXAMPLES.md)
2. **Learn Dataloaders**: See [Dataloader Guide](./DATALOADERS.md)
3. **Add Features**: Read [Development Guide](./DEVELOPMENT.md)
4. **Deploy**: Follow [Deployment Guide](./DEPLOYMENT.md)

## Getting Help

- **Documentation**: Check the `docs/` folder
- **Issues**: [GitHub Issues](https://github.com/ElioNeto/go-graphql-api-boilerplate/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ElioNeto/go-graphql-api-boilerplate/discussions)
- **gqlgen Docs**: [gqlgen.com](https://gqlgen.com/)

## Useful Commands

```bash
# Development
make run              # Run the API
make generate         # Generate GraphQL code
make tidy             # Tidy Go modules

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make migrate-create   # Create new migration

# Testing
make test             # Run tests
make test-coverage    # Run tests with coverage
make lint             # Run linters

# Building
make build            # Build binary
make docker-build     # Build Docker image

# Cleaning
make clean            # Remove build artifacts
```
