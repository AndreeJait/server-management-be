# Go Hex Boilerplate

A Go Hexagonal Architecture (Ports & Adapters) template repository.

Part of the ecosystem:

1. **[go-utility](https://github.com/AndreeJait/go-utility)** — shared infrastructure wrappers (logging, HTTP, DB, Redis, auth, storage, etc.)
2. **go-hex-boilerplate** (this repo) — hexagonal architecture project template
3. **andaliman** (coming soon) — CLI code generator for handlers, use cases, inbound/outbound ports, etc.

## Architecture

Strict inward dependency direction: **adapters → ports → domain**. Never the reverse.

```
cmd/
  http/                HTTP server entry point (wiring + DI)
  migrate/             Migration runner (up, down, fresh)
domain/                Core business logic (zero external dependencies)
  entity/              Domain models
  error/               Domain errors
port/                  Interface contracts
  inbound/             Driving ports (use case interfaces + input structs)
  outbound/            Driven ports (repository/service interfaces)
usecase/               Use case implementations (root-level, separate from ports)
adapter/               Concrete implementations of ports
  inbound/             HTTP handlers (echo/, gin/, mux/)
  outbound/            SQL, Redis, health, user repo implementations
config/                Configuration loading (Go code only)
files/                 Non-Go files
  config/              YAML configs (app.yaml + gitignored app.local.yaml)
  migrations/          SQL migration files
```

### Why `usecase/` is a root-level package

`port/inbound/` defines *what* the system can do (interfaces). `usecase/` implements *how* it does it (business logic). Separating them:

- **Clear separation** — contracts vs. implementations never mix in one package
- **No circular dependencies** — `usecase/` → `port/inbound/` → `domain/` is always one-directional
- **Generator-friendly** — andaliman scaffolds interface and implementation independently
- **Hex convention** — ports are the boundary, use cases are the application core

### Scaling: Sub-packages by Domain

Start flat, promote to a sub-package when a file gets unwieldy (1k+ lines):

```
# Small module
usecase/auth.go

# Large module — sub-package per operation
usecase/todo/create.go
usecase/todo/list.go
usecase/todo/update.go
usecase/todo/delete.go
```

## Multi-Engine HTTP Support

Three HTTP engines from [go-utility/v2](https://github.com/AndreeJait/go-utility), selectable at runtime:

| Engine | Flag | Framework |
|--------|------|-----------|
| Echo v5 (default) | `--engine=echo` | `labstack/echo/v5` |
| Gin | `--engine=gin` | `gin-gonic/gin` |
| Gorilla Mux | `--engine=mux` | `gorilla/mux` |

Each engine adapter lives in its own package under `adapter/inbound/`, keeping framework-specific code isolated. All three expose the same API.

## Authentication & RBAC

- **Register** — `POST /auth/register` (public) — creates user with bcrypt-hashed password, returns JWT
- **Login** — `POST /auth/login` (public) — verifies credentials, returns JWT
- **Global auth** — all other routes require a valid Bearer token (middleware validates via `authw.Authenticator`)
- **RBAC** — per-route permission checks using `RequirePermission(rbac, "todos:read")`

Default roles:

| Role | Permissions |
|------|------------|
| `admin` | `todos:read`, `todos:write`, `todos:delete`, `users:read` |
| `user` | `todos:read` |

## Dependency Injection (Uber dig)

DI uses [uber-go/dig](https://github.com/uber-go/dig). Provider functions are organized by layer:

| File | Providers |
|------|-----------|
| `cmd/http/infra.go` | DB, Redis, JWT, Authenticator, UserRepository, RBAC |
| `cmd/http/service.go` | HealthRepository, TodoRepository (with caching decorator), use cases |
| `cmd/http/router.go` | HTTP handler (engine selection) |
| `cmd/http/wire.go` | Container setup, `CleanupCollector` for graceful shutdown |

Adding a new dependency: write a provider function with its dependencies as parameters, then call `c.Provide(yourFunc)` in the appropriate `provide*` function. Dig resolves the rest.

## Getting Started

### Prerequisites

- Go 1.25+
- PostgreSQL
- Redis

### Setup

```bash
# Clone
git clone https://github.com/AndreeJait/go-hex-boilerplate.git
cd go-hex-boilerplate

# Copy local config and customize DSN/Redis address
cp files/config/app.local.yaml.example files/config/app.local.yaml

# Run migrations
make migrate-up

# Run with default engine (echo)
make run
```

### Run

```bash
make run                      # Default engine (echo)
make run-engine E=gin         # Specific engine (gin|mux)
make build                    # Build binary to bin/server
```

## Configuration

Config files in `files/config/`:

| File | Purpose |
|------|---------|
| `app.yaml` | Base config (committed) |
| `app.local.yaml` | Local overrides (gitignored) |
| `app.local.yaml.example` | Template — copy to `app.local.yaml` |

**Override priority** (highest wins): environment variables → `app.local.yaml` → `app.yaml`

```yaml
app:
  name: go-hex-boilerplate
  env: development
  http_port: 8080

http:
  engine: echo              # echo | gin | mux
  enable_swagger: true
  debug_mode: true

log:
  level: debug              # debug | info | warn | error
  format: JSON              # JSON | TEXT

db:
  driver: gorm              # gorm | sqlx
  dialect: postgres
  dsn: "postgres://user:pass@localhost:5432/go_hex_boilerplate?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  debug_mode: false

redis:
  address: "localhost:6379"
  password: ""
  db: 0
  pool_size: 10

auth:
  jwt_secret: "change-me-in-production"
  jwt_ttl: 24h
  jwt_issuer: "go-hex-boilerplate"

graceful:
  shutdown_timeout: 10s
```

Environment variable overrides: `APP_HTTP_PORT=9090`, `HTTP_ENGINE=gin`, `DB_DSN=postgres://...`, `AUTH_JWT_SECRET=...`, etc.

## Database Migrations

```bash
make migrate-new name=create_users_table   # Create new migration
make migrate-up                            # Run pending migrations
make migrate-down                          # Roll back last migration
make migrate-fresh                         # Drop all + re-run all
```

Migration files are in `files/migrations/` — timestamped `.up.sql` and `.down.sql` pairs.

## Redis & Caching

Redis is an **infrastructure concern**, not a domain port. Caching uses the Decorator pattern:

- `todoOutbound.NewCachingRepository(baseRepo, redisClient)` wraps `TodoRepository` with Redis caching
- The domain only knows about `port/outbound.TodoRepository` — it never sees Redis
- If Redis is down, the caching decorator falls through to the database

## Swagger

Swagger annotations are written only on Echo handlers (the default engine). All three engines share the same API.

```bash
make swag    # Generate docs from Echo annotations (auto-installs swag if missing)
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run with default engine (echo) |
| `make run-engine E=gin` | Run with specific engine |
| `make build` | Build binary to `bin/server` |
| `make swag` | Generate Swagger docs |
| `make test` | Run all tests |
| `make vet` | Run static analysis |
| `make tidy` | Clean up dependencies |
| `make migrate-new name=foo` | Create new migration |
| `make migrate-up` | Run pending migrations |
| `make migrate-down` | Roll back last migration |
| `make migrate-fresh` | Drop all + re-run all |

## License

MIT