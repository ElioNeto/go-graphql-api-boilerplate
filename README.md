# Go GraphQL API Boilerplate

A production-ready GraphQL API built with Go, **gqlgen** (schema-first), **PostgreSQL (sqlx)**, **JWT Authentication**, and **Dataloaders** to solve N+1 problems.

[![CI](https://github.com/ElioNeto/go-graphql-api-boilerplate/actions/workflows/ci.yml/badge.svg)](https://github.com/ElioNeto/go-graphql-api-boilerplate/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/go-1.22-blue.svg)](https://go.dev/)

## Architecture

```mermaid
graph TD
    Client([GraphQL Client])
    subgraph API Server
        MW[Middleware<br>Logging · JWT · Dataloader]
        GQ[GraphQL Handler<br>gqlgen]
        RES[Resolvers<br>graph/resolvers]
        SVC[Services<br>internal/services]
        REPO[Repositories<br>internal/repositories]
    end
    DB[(PostgreSQL)]

    Client -->|HTTP POST /query| MW
    MW --> GQ
    GQ --> RES
    RES -->|Call| SVC
    SVC -->|Call| REPO
    REPO -->|Query| DB
```

## Directory Structure
- `cmd/api/main.go` - Application entry point. Wires DB, Config, Handlers, and Middlewares.
- `graph/` - gqlgen related folder.
  - `schema.graphqls` - GraphQL schema definition.
  - `schema.resolvers.go` - The actual resolver implementation connecting to `internal/services`.
- `internal/`
  - `config/` - Environment variable parsing via Viper.
  - `dataloaders/` - `dataloadgen` setup to resolve N+1 queries.
  - `middleware/` - Auth, Logging, and Dataloader injection.
  - `models/` - Domain Models mapped natively to GraphQL via `autobind`.
  - `repositories/` - Data layer interfaces & implementations (PostgreSQL).
  - `services/` - Pure business logic interfaces & implementations.
- `migrations/` - SQL migration files applied automatically on startup.

## Quick Start

### 1. Configure
```bash
git clone https://github.com/ElioNeto/go-graphql-api-boilerplate.git
cd go-graphql-api-boilerplate
cp .env.example .env
```

### 2. Generate GraphQL Code
Since we are using `gqlgen`, you must generate the code after schema changes:
```bash
make generate
```

### 3. Database & Run
```bash
createdb graphql_db
make migrate-up
make run
```
Go to `http://localhost:8080/` to access the GraphQL Playground.

## GraphQL Usage

### Register a User
```graphql
mutation {
  createUser(input: {
    name: "Alice",
    email: "alice@example.com",
    password: "password123"
  }) {
    id
    name
    email
  }
}
```

### Login
```graphql
mutation {
  login(input: {
    email: "alice@example.com",
    password: "password123"
  }) {
    token
    user {
      id
      name
    }
  }
}
```

### Get Current User (Authenticated)
Add `Authorization: Bearer <token>` in the HTTP headers of your request.
```graphql
query {
  me {
    id
    name
    email
  }
}
```
