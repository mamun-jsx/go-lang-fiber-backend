# Request Workflow

This document explains **what happens from the moment a client sends a request until it gets a response**.

## The big picture

```
 ┌────────┐   HTTP    ┌────────────┐   ┌─────────┐   ┌─────────┐   ┌────────────┐   ┌──────────┐
 │ Client │ ───────▶ │ Middleware │ ─▶│ Router  │ ─▶│ Handler │ ─▶│  Service   │ ─▶│Repository│ ─▶ PostgreSQL
 └────────┘           └────────────┘   └─────────┘   └─────────┘   └────────────┘   └──────────┘
      ▲                                                  │               │                │
      │                                                  ▼               ▼                ▼
      └────────────── JSON response ◀── DTO ◀──────── Handler ◀──── Service ◀──────── Model
```

**Rule:** each layer only talks to the layer directly below it.
`Handler → Service → Repository → Database`, never `Handler → Database`.

---

## Step 0 – Application startup (`cmd/api/main.go`)

Before any request comes in, `main.go` builds the chain of layers once:

```go
cfg, _         := config.Load()                                   // 1. read .env / env vars
db, _          := database.Connect(cfg.DB, cfg.App.Env)           // 2. connect to PostgreSQL
database.Migrate(db)                                              // 3. create "products" table
productRepo    := repository.NewProductRepository(db)             // 4. repository gets the DB
productService := service.NewProductService(productRepo)          // 5. service gets the repository
productHandler := handler.NewProductHandler(productService, v)    // 6. handler gets the service
router.Setup(app, productHandler)                                 // 7. routes point to handlers
app.Listen(":8080")                                               // 8. start listening
```

This is **dependency injection**: each layer receives what it needs, so nothing is global and everything is testable.

---

## Step-by-step: `POST /api/v1/products` (Create)

Client sends:
```http
POST /api/v1/products
Content-Type: application/json

{ "name": "Laptop", "price": 1299.5, "short_description": "14 inch ultrabook" }
```

### 1. Middleware (`internal/router/router.go`)
The request passes through global middleware, in order:
1. `recover` – if anything panics, it turns into a 500 instead of crashing the server.
2. `requestid` – adds an `X-Request-ID` header for tracing.
3. `logger` – logs method, path, status and latency.
4. `cors` – adds CORS headers.

### 2. Router (`internal/router/router.go`)
The router matches the method + path and calls the handler:
```go
products.Post("/", productHandler.Create)
```

### 3. Handler (`internal/handler/product_handler.go`) — *HTTP layer*
```go
var req dto.CreateProductRequest
c.Bind().JSON(&req)                         // a) parse JSON  -> 400 if malformed
h.validator.Validate(req)                   // b) validate    -> 422 if invalid
product, err := h.service.Create(ctx, req)  // c) call the SERVICE
return response.Success(c, 201, "...", product) // d) send JSON response
```
The handler only deals with HTTP things: reading input, validation, status codes.

### 4. Service (`internal/service/product_service.go`) — *business logic layer*
```go
p := &model.Product{ Name: strings.TrimSpace(req.Name), ... } // a) DTO -> Model (+ business rules)
s.repo.Create(ctx, p)                                         // b) call the REPOSITORY
return dto.ToProductResponse(p)                               // c) Model -> Response DTO
```
The service knows nothing about HTTP. It applies the rules (trim input, "at least one field on update",
translate "not found" into a 404 `AppError`, etc.).

### 5. Repository (`internal/repository/product_repository.go`) — *data layer*
```go
r.db.WithContext(ctx).Create(p)   // GORM generates the SQL below
```
```sql
INSERT INTO products (name, price, short_description, created_at, updated_at, deleted_at)
VALUES ('Laptop', 1299.5, '14 inch ultrabook', NOW(), NOW(), NULL) RETURNING id;
```
The repository is the **only** place that talks to the database.

### 6. PostgreSQL
Stores the row and returns the new `id`. GORM fills `p.ID`, `p.CreatedAt`, `p.UpdatedAt`.

### 7. Response goes back up
`Repository → Service (Model → DTO) → Handler → JSON`:
```json
HTTP/1.1 201 Created
{
  "success": true,
  "message": "product created successfully",
  "data": {
    "id": 1,
    "name": "Laptop",
    "price": 1299.5,
    "short_description": "14 inch ultrabook",
    "created_at": "2026-09-23T17:25:21Z",
    "updated_at": "2026-09-23T17:25:21Z"
  }
}
```

---

## How errors flow

Every layer simply **returns** an error; nobody writes error JSON by hand.

```
Repository            Service                         Handler          Global ErrorHandler
──────────            ───────                         ───────          ───────────────────
ErrNotFound   ──▶   apperror.NotFound(...)     ──▶  return err   ──▶  404 {"success":false,"message":"product not found"}
DB error      ──▶   apperror.Internal(err)     ──▶  return err   ──▶  500 {"success":false,"message":"internal server error"}  (real error is logged)
                                     bad JSON  ──▶  apperror.BadRequest   ──▶ 400
                              validation fail  ──▶  apperror.Validation   ──▶ 422 {"errors":[{"field":"price","message":"price must be greater than 0"}]}
```

`internal/middleware/error_handler.go` is registered in `fiber.Config{ErrorHandler: ...}` and converts
**every** returned error into the same JSON format. Internal details are logged, never sent to the client.

---

## The other endpoints (same flow, different methods)

| Request | Handler | Service | Repository | SQL (generated by GORM) |
|---------|---------|---------|------------|-------------------------|
| `GET /api/v1/products?page=1&limit=10` | `List` binds query | `List` normalizes page/limit, builds meta | `FindAll` | `SELECT count(*) ...; SELECT * FROM products WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 10 OFFSET 0` |
| `GET /api/v1/products/:id` | `GetByID` parses id | `GetByID` → 404 if missing | `FindByID` | `SELECT * FROM products WHERE id = ? AND deleted_at IS NULL LIMIT 1` |
| `PUT /api/v1/products/:id` | `Update` parses id + body | `Update` loads product, applies only sent fields | `FindByID` + `Update` | `UPDATE products SET name=?, price=?, ... WHERE id = ?` |
| `DELETE /api/v1/products/:id` | `Delete` parses id | `Delete` → 404 if missing | `Delete` | `UPDATE products SET deleted_at = NOW() WHERE id = ?` (soft delete) |

---

## Summary

1. **Client** sends an HTTP request.
2. **Middleware** handles cross-cutting concerns (panic recovery, request ID, logging, CORS).
3. **Router** picks the right handler.
4. **Handler** parses and validates input, then calls the service.
5. **Service** applies business rules, then calls the repository.
6. **Repository** runs the SQL through GORM against **PostgreSQL**.
7. The result travels back up, gets converted **Model → DTO**, and the handler returns **JSON**.
8. Any error at any layer is returned upward and formatted by the **global error handler**.
