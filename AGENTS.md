# AGENTS.md

## Project Overview

A production-ready Go REST API boilerplate following **Clean Architecture** principles. It provides user CRUD, JWT authentication (access + refresh tokens with HttpOnly cookies), pagination, sorting, and search out of the box.

**Architecture layers** (dependencies point inward):

```
Delivery (HTTP handlers, middleware, DTOs)
    → Usecase (business logic)
        → Domain (entities, interfaces — zero external deps)
    → Repository (data access via GORM/PostgreSQL)
```

**Tech stack:** Go 1.24 · Fiber v2 · GORM · PostgreSQL 16 · JWT (`golang-jwt/v5`) · Viper · zerolog · `go-playground/validator` · Swaggo · testify

## Setup

### Prerequisites

- Go 1.22+ (module requires 1.24)
- PostgreSQL 16 (or Docker)

### Quick start

```bash
# 1. Clone and configure
git clone https://github.com/azbagas/golang-clean-arch-boilerplate.git
cd golang-clean-arch-boilerplate
cp .env.example .env
# Edit .env with your database credentials and JWT secret

# 2. Start PostgreSQL via Docker
docker-compose up -d postgres

# 3. Install deps and run
go mod tidy
make run
```

The API starts at `http://localhost:8080`. Swagger docs at `http://localhost:8080/swagger/`.

## Build & Dev Commands

| Command              | Description                                      |
|----------------------|--------------------------------------------------|
| `make run`           | Run the application (`go run cmd/api/main.go`)   |
| `make build`         | Build binary to `bin/server`                     |
| `make test`          | Run all tests with verbose output                |
| `make test-cover`    | Run tests and generate `coverage.html`           |
| `make lint`          | Run `go vet ./...`                               |
| `make swagger`       | Regenerate Swagger docs (requires `swag` CLI)    |
| `make migrate-up`    | Apply database migrations                        |
| `make migrate-down`  | Rollback database migrations                     |
| `make docker-up`     | Start API + PostgreSQL containers                |
| `make docker-down`   | Stop Docker containers                           |
| `make docker-logs`   | Tail container logs                              |
| `make clean`         | Remove `bin/`, `tmp/`, coverage files            |

### Tool installation

```bash
# Swagger doc generator
go install github.com/swaggo/swag/cmd/swag@latest

# Database migrations CLI
# See: https://github.com/golang-migrate/migrate
```

## Testing

### Run all tests

```bash
make test
# or directly:
go test -v ./...
```

### Run tests for a specific package

```bash
go test -v ./internal/usecase/...
go test -v ./internal/delivery/http/handler/...
```

### Run a single test function

```bash
go test -v -run TestUserUsecase_Create ./internal/usecase/...
go test -v -run TestUserUsecase_Create/success ./internal/usecase/...
```

### Test coverage

```bash
make test-cover
# Opens coverage.html with line-by-line report
```

### Testing patterns

- **Framework:** `testify` (assert + mock)
- **Mocks:** Hand-written in `internal/mocks/mock_repository.go`, implementing domain interfaces using `testify/mock`
- **Test file placement:** `_test.go` files live next to the code they test, using `_test` package suffix (e.g., `package usecase_test`)
- **Test structure:** Table-driven subtests using `t.Run("case name", ...)` with shared mock setup per function
- **Assertions:** `assert.NoError`, `assert.ErrorIs`, `assert.Equal`, `assert.Nil`
- **Mock lifecycle:** Use `.Once()` per expectation and `AssertExpectations(t)` at the end of each subtest

### Current test coverage

Tests exist for:
- `internal/usecase/` — `user_usecase_test.go`, `auth_usecase_test.go`
- `internal/delivery/http/handler/` — `user_handler_test.go`, `auth_handler_test.go`

No tests for: repositories (require database), middleware, `pkg/` packages.

## Code Style & Conventions

### General

- **Indentation:** Tabs (standard `gofmt`)
- **Line endings:** CRLF (Windows)
- **Formatter:** `gofmt` / `go vet`
- **No linter config files** — only `go vet` is used via `make lint`

### Naming

- **Packages:** lowercase single words (`domain`, `usecase`, `handler`, `dto`, `mocks`)
- **Interfaces:** noun describing the role — `UserRepository`, `UserUsecase`, `AuthUsecase`
- **Structs implementing interfaces:** same name without the "I" prefix — `userUsecase`, `userRepository`
- **Constructors:** `NewXxx()` pattern — `NewUserHandler()`, `NewUserUsecase()`, `NewRouter()`
- **Mock types:** `MockXxxYyy` — `MockUserRepository`, `MockAuthUsecase`
- **Handler methods:** HTTP verb / CRUD name — `Create`, `GetAll`, `GetByID`, `Update`, `Delete`
- **DTOs:** `XxxRequest` / `XxxResponse` — `CreateUserRequest`, `UserResponse`
- **Sentinel errors:** `Err` prefix — `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`

### Patterns

- **Dependency injection:** Manual, wired in `cmd/api/main.go` (no DI framework)
- **Config:** Viper reads `.env` file + environment variables; structured into `Config` > `ServerConfig`, `DatabaseConfig`, `JWTConfig`, `LoggerConfig`
- **Error handling:** Domain sentinel errors (`internal/domain/errors.go`) mapped to HTTP status codes in `pkg/response/response.go`
- **Request/Response:** DTOs in `internal/delivery/http/dto/` separate API schema from domain entities
- **Validation:** `go-playground/validator` struct tags with manual validation in handlers
- **Swagger annotations:** Godoc-style comments on handler methods (e.g., `// @Summary`, `// @Router`)
- **Logging:** `zerolog` logger initialized in `main.go`, passed to middleware
- **Graceful shutdown:** Signal handling with `os.Signal` channel in `main.go`

### API conventions

- **Base path:** `/api/v1`
- **Route groups:** `/api/v1/auth` (public + protected), `/api/v1/users` (protected)
- **Response format:** `{ "success": bool, "data": ... }` or `{ "success": bool, "error": "..." }`
- **Paginated responses:** Include `pagination` object with `page`, `per_page`, `total_items`, `total_pages`, `has_next`, `has_prev`

## Architecture Notes

### Directory structure

```
cmd/api/main.go              → Entry point, DI wiring, server startup
internal/
  domain/                    → Core entities, interfaces, sentinel errors, query params
    user.go                  → User, RefreshToken, TokenPair entities + all interfaces
    errors.go                → Sentinel errors (ErrNotFound, ErrConflict, etc.)
    query.go                 → PaginationParams, SortParams, PaginatedResult
  usecase/                   → Business logic implementations
    user_usecase.go          → UserUsecase implementation
    auth_usecase.go          → AuthUsecase (register, login, refresh, logout)
  repository/postgres/       → GORM-based PostgreSQL repositories
    user_repository.go       → UserRepository implementation
    refresh_token_repository.go → RefreshTokenRepository implementation
  delivery/http/             → HTTP layer
    router.go                → Route registration, middleware setup
    handler/                 → Request handlers (auth_handler.go, user_handler.go)
    dto/                     → Request/response data transfer objects
    middleware/              → CORS, JWT auth, request logger
  mocks/                     → Hand-written testify mocks
    mock_repository.go       → Mocks for all domain interfaces
pkg/                         → Shared infrastructure (importable by any layer)
  config/config.go           → Viper-based config loader
  database/postgres.go       → GORM database connection
  logger/logger.go           → zerolog logger setup
  response/response.go       → Standardized HTTP response helpers
migrations/                  → SQL migration files (golang-migrate format)
docs/                        → Auto-generated Swagger docs (do not edit manually)
```

### Key design decisions

1. **All interfaces live in `internal/domain/`** — the innermost layer owns contracts
2. **`pkg/response`** imports `internal/domain` to map errors → HTTP statuses (boundary crossing kept minimal)
3. **GORM AutoMigrate** runs on startup alongside the `migrations/` directory SQL files
4. **Refresh tokens** are stored in PostgreSQL, rotated on each refresh, and delivered via HttpOnly cookies

### Adding a new entity

1. Define the entity struct and repository/usecase interfaces in `internal/domain/`
2. Implement the repository in `internal/repository/postgres/`
3. Implement the usecase in `internal/usecase/`
4. Create DTOs in `internal/delivery/http/dto/`
5. Create a handler in `internal/delivery/http/handler/`
6. Add mocks in `internal/mocks/`
7. Register routes in `internal/delivery/http/router.go`
8. Wire dependencies in `cmd/api/main.go`

## Security Considerations

### Files that must never be committed

- **`.env`** — Contains `JWT_SECRET`, database credentials. Listed in `.gitignore`.
- **`coverage.out`, `coverage.html`** — May reveal internal code paths.

### Sensitive areas

- **`JWT_SECRET`** — Used by `pkg/config/` and the auth usecase. Must be a strong, unique value in production.
- **`cmd/api/main.go`** — Wires all security-sensitive components; changes here affect the entire app.
- **`internal/delivery/http/middleware/jwt.go`** — JWT validation middleware; errors here break auth.
- **`internal/usecase/auth_usecase.go`** — Password hashing (bcrypt), token generation, refresh token rotation.
- **`docker-compose.yml`** — Contains default database credentials; do not use defaults in production.
- **`Makefile` migrate targets** — Contain hardcoded database connection strings.

### Security features to preserve

- Refresh token stored as **HttpOnly, Secure (prod), SameSite=Lax** cookie
- **Token rotation** on refresh (old token deleted)
- Passwords hashed with **bcrypt**
- CORS middleware configured in `middleware/cors.go`

## Agent-Specific Notes

### Gotchas

- **`docs/` is auto-generated.** Never edit `docs/docs.go`, `docs/swagger.json`, or `docs/swagger.yaml` by hand. Run `make swagger` to regenerate after changing handler annotations.
- **GORM AutoMigrate vs SQL migrations:** Both exist. AutoMigrate runs on every startup. SQL migrations in `migrations/` are the canonical schema source — keep them in sync.
- **Mock file is hand-written.** When adding a new interface method to `domain/`, you must manually update `internal/mocks/mock_repository.go`.
- **Test packages use `_test` suffix** (e.g., `package usecase_test`), meaning only exported symbols are testable.
- **No CI/CD workflows** — No `.github/workflows/` directory exists. Tests must be run locally.

### Tips

- Always run `make test` after changes to verify nothing is broken.
- After modifying handler Swagger annotations, run `make swagger` and verify at `/swagger/`.
- When modifying domain interfaces, update mocks → usecase tests → handler tests in that order.
- The `pkg/response` package centralizes error-to-HTTP-status mapping — add new domain errors there.
- Use `go vet ./...` (or `make lint`) before committing.
