# Project Structure

A **Product CRUD REST API** built with **Go**, **Fiber v3**, **GORM** and **PostgreSQL**,
following the widely used *Clean / Layered Architecture* (Handler → Service → Repository).

```
go-lang-fiber-backend/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point: loads config, connects DB, wires layers, starts server
│
├── internal/                          # Private application code (cannot be imported by other modules)
│   ├── config/
│   │   └── config.go                  # Reads env vars / .env into a typed Config struct
│   ├── database/
│   │   └── postgres.go                # Opens GORM + PostgreSQL connection, pool settings, AutoMigrate
│   ├── model/
│   │   └── product.go                 # GORM entity -> "products" table
│   ├── dto/
│   │   └── product_dto.go             # Request/Response shapes, pagination, model <-> DTO mappers
│   ├── apperror/
│   │   └── errors.go                  # Typed AppError (404, 400, 422, 500) used across layers
│   ├── repository/
│   │   └── product_repository.go      # Data-access layer: the ONLY place that runs SQL via GORM
│   ├── service/
│   │   ├── product_service.go         # Business logic layer
│   │   └── product_service_test.go    # Unit tests using an in-memory fake repository
│   ├── handler/
│   │   └── product_handler.go         # HTTP layer: parse/validate request, call service, send response
│   ├── middleware/
│   │   └── error_handler.go           # Global error handler -> consistent JSON error responses
│   └── router/
│       └── router.go                  # Registers global middleware + /api/v1/products routes
│
├── pkg/                               # Reusable, app-agnostic helpers (safe to import elsewhere)
│   ├── response/
│   │   └── response.go                # Standard JSON envelope {success, message, data, meta, errors}
│   └── validator/
│       └── validator.go               # go-playground/validator wrapper with friendly field errors
│
├── .env.example                       # Template for environment variables (copy to .env)
├── .gitignore                         # Ignores .env, binaries, IDE files
├── go.mod / go.sum                    # Go module definition and dependency checksums
├── PROJECT_STRUCTURE.md               # This file
└── WORKFLOW.md                        # Request lifecycle: Request -> Handler -> Service -> Repository
```

## Responsibility of each layer

| Layer | Folder | Knows about | Must NOT do |
|-------|--------|-------------|-------------|
| **Entry point** | `cmd/api` | Everything (wiring only) | Contain business logic |
| **Router** | `internal/router` | Handlers, middleware | Contain logic |
| **Handler** | `internal/handler` | Fiber, DTOs, Service interface | Touch the database |
| **Service** | `internal/service` | DTOs, Models, Repository interface | Know about HTTP / Fiber |
| **Repository** | `internal/repository` | GORM, Models | Contain business rules |
| **Model** | `internal/model` | GORM tags | Be sent directly to clients |
| **DTO** | `internal/dto` | JSON / validation tags | Contain DB logic |

### Why this structure is "industry standard"
- **`cmd/` + `internal/` + `pkg/`** – the common Go project layout.
- **Dependency Injection** – each layer receives its dependency via a constructor (`NewXxx`) in `main.go`.
- **Interfaces between layers** – `ProductService` and `ProductRepository` are interfaces, so each layer can be unit-tested with fakes (see `product_service_test.go`).
- **DTOs separate from Models** – the database schema never leaks into the public API.
- **Centralized error handling** – layers return typed `AppError`s; one middleware converts them to JSON.
- **Consistent response envelope** – every response has the same shape.
- **Config from environment** – 12-factor style; no secrets in code.
- **Production concerns** – connection pooling, request timeouts, panic recovery, request IDs, request logging, CORS, soft delete, pagination, graceful shutdown.

## Product entity

| Field | Type | Rules |
|-------|------|-------|
| `id` | uint | auto-increment primary key |
| `name` | string | required, 2–255 chars |
| `price` | float (numeric(12,2)) | required, > 0 |
| `short_description` | string | optional, max 500 chars |
| `created_at` / `updated_at` | timestamp | set automatically |
| `deleted_at` | timestamp | soft delete (hidden from API) |

## API endpoints

| Method | URL | Description | Success |
|--------|-----|-------------|---------|
| GET | `/health` | Health check | 200 |
| POST | `/api/v1/products` | Create a product | 201 |
| GET | `/api/v1/products?page=1&limit=10` | List products (paginated, max limit 100) | 200 |
| GET | `/api/v1/products/:id` | Get one product | 200 |
| PUT | `/api/v1/products/:id` | Update product (send only fields to change) | 200 |
| DELETE | `/api/v1/products/:id` | Soft-delete product | 200 |

Error codes: `400` bad JSON / bad id / empty update, `404` not found, `422` validation failed, `500` internal error.

## Getting started

### Requirements
- Go (version in `go.mod`)
- A running PostgreSQL server

### Steps
```bash
# 1. Create the database
createdb -U postgres product_db

# 2. Create your env file and edit the DB credentials
cp .env.example .env

# 3. Download dependencies
go mod download

# 4. Run the server (tables are created automatically on startup)
go run ./cmd/api

# 5. Run the tests
go test ./...
```

### Example requests
```bash
# Create
curl -X POST http://localhost:8080/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":1299.50,"short_description":"14 inch ultrabook"}'

# List
curl "http://localhost:8080/api/v1/products?page=1&limit=10"

# Get one
curl http://localhost:8080/api/v1/products/1

# Update (partial)
curl -X PUT http://localhost:8080/api/v1/products/1 \
  -H "Content-Type: application/json" \
  -d '{"price":1199}'

# Delete
curl -X DELETE http://localhost:8080/api/v1/products/1
```
