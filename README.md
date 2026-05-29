# go-template

A production-ready Go service skeleton for building HTTP APIs quickly and correctly. It comes pre-wired with Postgres, structured logging, Prometheus metrics, database migrations, and a clean layered architecture — so you can start writing domain logic on day one instead of plumbing.

## What's included

| Concern | Library |
|---|---|
| HTTP router | [go-chi/chi v5](https://github.com/go-chi/chi) |
| Database driver | [jackc/pgx v5](https://github.com/jackc/pgx) + [jmoiron/sqlx](https://github.com/jmoiron/sqlx) |
| Migrations | [golang-migrate/migrate v4](https://github.com/golang-migrate/migrate) |
| Logger | stdlib `log/slog` (JSON or text, configured via `LOG_FORMAT`) |
| Metrics | [prometheus/client_golang](https://github.com/prometheus/client_golang) |
| Config | [spf13/viper](https://github.com/spf13/viper) (reads `.env`) |
| Mocks | [vektra/mockery](https://github.com/vektra/mockery) |
| Linter | [golangci-lint](https://golangci-lint.run/) |

## Quick start

```sh
# 1. Install dev tools into ./bin/
make tools

# 2. Vendor dependencies
make vendor

# 3. Configure environment
cp sample.env .env

# 4. Start Postgres
docker-compose up -d postgres

# 5. Apply migrations
make migrate

# 6. Build and run
make build
./bin/service
```

The service starts on `http://localhost:8080`.

**Endpoints out of the box:**

| Method | Path | Description |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/metrics` | Prometheus metrics |
| GET | `/example` | Example user endpoint |

## Running with Docker

```sh
docker-compose up -d
```

This builds the service image and starts it alongside Postgres. The service waits for Postgres to be healthy before starting.

## Project structure

```
go-template/
├── cmd/
│   ├── service/          # Main HTTP server binary
│   ├── worker/           # Background worker binary
│   ├── fan_out/          # Fan-out concurrency example
│   └── shared_memory/    # Shared memory concurrency example
│
├── internal/
│   ├── app/
│   │   └── http_server/  # Composition root: wires repos, usecases, routes
│   ├── connectors/
│   │   ├── logger/       # slog setup (text/JSON handler, level parsing)
│   │   ├── metrics/      # Prometheus registry
│   │   └── postgres/     # sqlx connection setup
│   ├── env/              # Config struct, Viper reader
│   ├── http/
│   │   └── handlers/     # HTTP handler structs (one folder per endpoint)
│   ├── pkg/
│   │   └── users/        # Example domain package
│   │       ├── entity/   # Domain structs
│   │       ├── usecase/  # Business logic implementation
│   │       ├── repository/ # Postgres queries
│   │       ├── mocks/    # Auto-generated mocks
│   │       ├── usecase.go  # Usecase interface
│   │       └── repository.go # Repository interface
│   └── utils/
│
├── db/
│   └── migrations/       # SQL migration files (up/down)
├── docs/
│   ├── structure.md      # Project layout conventions
│   └── code_style.md     # Code style rules
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── sample.env
```

## Adding a new domain

1. Create `internal/pkg/<your_domain>/` following the `users` package as a reference:
   - Define `Usecase` and `Repository` interfaces with `//go:generate` directives
   - Implement them in `usecase/usecase.go` and `repository/repository.go`
   - Put domain structs in `entity/`

2. Register the new repo and usecase in `internal/app/http_server/repository.go` and `usecase.go`.

3. Add a handler in `internal/http/handlers/<your_handler>/` and register the route in `internal/app/http_server/routes.go`.

4. For external API clients, create `internal/clients/<client_name>/client.go`.

## Migrations

```sh
make migrate                      # apply all pending migrations
make migrate-down                 # roll back all migrations
make migrate-step-down            # roll back one migration
make migrate-version              # print current migration version
make migrate-force v=<version>    # force-set version (recover dirty state)
make migrate-create name=<name>   # create a new sequential migration pair
make migrate-drop                 # drop everything in the database
```

## Development

```sh
make full-lint          # goimports + fmt + vet + golangci-lint
go test ./...           # all tests
go generate ./internal/pkg/<domain>/...   # regenerate mocks
```

## Docs

- [Project layout conventions](./docs/structure.md)
- [Code style guide](./docs/code_style.md)
