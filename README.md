# 🏗️ Golang Clean Architecture Boilerplate

A **production-ready** Go REST API boilerplate following Clean Architecture principles. Features user CRUD operations, paginated listing with sorting and search, JWT authentication (access + refresh tokens via HttpOnly cookies), and auto-generated Swagger docs. Built with Fiber, GORM, and PostgreSQL.

## 🏛️ Architecture

```
┌─────────────────────────────────────────────────┐
│                  Delivery Layer                 │
│         (HTTP Handlers, Middleware, DTOs)       │
├─────────────────────────────────────────────────┤
│                   Usecase Layer                 │
│                 (Business Logic)                │
├─────────────────────────────────────────────────┤
│                   Domain Layer                  │ 
│              (Entities, Interfaces)             │
├─────────────────────────────────────────────────┤
│                 Repository Layer                │
│          (Data Access, PostgreSQL/GORM)         │
└─────────────────────────────────────────────────┘
```

**Dependency Rule**: Dependencies always point **inward**. The domain layer has zero external dependencies.

## 🛠️ Tech Stack

| Category | Technology |
|---|---|
| Web Framework | Fiber v2 |
| ORM | GORM |
| Database | PostgreSQL |
| Auth | JWT access tokens + HttpOnly refresh token cookies |
| Config | Viper |
| Logger | zerolog |
| Validation | go-playground/validator |
| Docs | Swaggo (Swagger) |
| Testing | testify (assert + mock) |

## 📁 Project Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/           # Entities + interfaces (innermost layer)
│   ├── usecase/          # Business logic
│   ├── repository/       # Data access (GORM/PostgreSQL)
│   ├── delivery/http/    # Handlers, middleware, DTOs, router
│   └── mocks/            # Test mocks
├── pkg/                  # Shared infrastructure (config, db, logger)
├── migrations/           # SQL migration files
├── docs/                 # Swagger generated docs
├── Dockerfile            # Multi-stage Docker build
├── docker-compose.yml    # API + PostgreSQL
└── Makefile              # Dev commands
```

## 🚀 Quick Start

### Prerequisites
- Go 1.22+
- PostgreSQL (or Docker)

### 1. Clone & configure

```bash
git clone https://github.com/azbagas/golang-clean-arch-boilerplate.git
cd golang-clean-arch-boilerplate
cp .env.example .env
# Edit .env with your database credentials
```

### 2. Start PostgreSQL (Docker)

```bash
docker-compose up -d postgres
```

### 3. Run the API

```bash
go mod tidy
go run cmd/api/main.go
```

Or using Make:

```bash
make run
```

The API will be available at `http://localhost:8080`.

## 🔐 Authentication Flow

The API uses **short-lived access tokens** (JWT, 15 min) and **long-lived refresh tokens** (opaque, 7 days) stored in the database and delivered via **HttpOnly cookies**.

```
1. POST /api/v1/auth/register  →  Create account
2. POST /api/v1/auth/login     →  Get access_token (JSON) + refresh_token (HttpOnly cookie)
3. Use access_token in Authorization: Bearer <token> header
4. GET  /api/v1/auth/current   →  Get current user profile
5. POST /api/v1/auth/refresh   →  Exchange cookie for new token pair (rotation)
6. POST /api/v1/auth/logout    →  Revoke refresh token + clear cookie
```

**Security features:**
- Refresh token is **HttpOnly, Secure (prod), SameSite=Lax** — never exposed to JavaScript
- **Token rotation** on every refresh (old token deleted, new one issued)
- **Single-session logout** (only the current refresh token is revoked)
- Passwords hashed with **bcrypt**

## 📡 API Endpoints

### Auth (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/login` | Login → access token + refresh cookie |
| POST | `/api/v1/auth/refresh` | Refresh token pair (cookie-based) |

### Auth (Protected - JWT Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/auth/current` | Get current user profile |
| POST | `/api/v1/auth/logout` | Logout → revoke refresh token |

### Users (Protected - JWT Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users` | Create a user |
| GET | `/api/v1/users` | Get all users (paginated, sortable, searchable) |
| GET | `/api/v1/users/:id` | Get user by ID |
| PUT | `/api/v1/users/:id` | Update a user |
| DELETE | `/api/v1/users/:id` | Delete a user |

#### Query Parameters for `GET /api/v1/users`

| Parameter | Default | Description |
|-----------|---------|-------------|
| `page` | `1` | Page number (≥ 1) |
| `page_size` | `10` | Items per page (1–100) |
| `sort_by` | — | Sort field: `name`, `email`, `created_at` |
| `sort_order` | `asc` | Sort direction: `asc` or `desc` |
| `search` | — | Search by name or email (case-insensitive) |

**Example:**

```
GET /api/v1/users?page=1&page_size=10&sort_by=name&sort_order=desc&search=john
```

### Swagger UI

Visit `http://localhost:8080/swagger/` for interactive API documentation.

## 🧪 Testing

```bash
# Run all tests
make test

# Run with coverage
make test-cover
```

## 🐳 Docker

```bash
# Start everything (API + PostgreSQL)
make docker-up

# Stop
make docker-down

# View logs
make docker-logs
```

## ⚙️ Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | API server port |
| `SERVER_MODE` | `development` | `development` or `production` |
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `clean_arch_db` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode |
| `JWT_SECRET` | — | JWT signing secret |
| `JWT_ACCESS_EXPIRY` | `15m` | Access token lifetime |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh token lifetime (7 days) |
| `LOG_LEVEL` | `debug` | Log level |

## 📝 Available Make Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the application |
| `make build` | Build binary |
| `make test` | Run tests |
| `make test-cover` | Run tests with coverage report |
| `make lint` | Run go vet |
| `make swagger` | Generate Swagger docs |
| `make migrate-up` | Run database migrations |
| `make migrate-down` | Rollback migrations |
| `make docker-up` | Start Docker containers |
| `make docker-down` | Stop Docker containers |
| `make clean` | Clean build artifacts |

## 📄 License

Licensed under the [MIT License](LICENSE).
