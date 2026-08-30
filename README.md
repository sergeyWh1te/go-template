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
| Linter | [golangci-lint](https://golangci-lint.run/) v2 |
| Vulnerabilities | [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) |

## Quick start

```sh
# 1. Install dev tools into ./bin/
make tools

# 2. Configure environment
cp sample.env .env

# 3. Start Postgres
docker compose up -d postgres

# 4. Apply migrations
make migrate

# 5. Build and run
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
make up        # start postgres + service
make logs      # follow the logs
make down      # stop the stack
```

This builds the service image and starts it alongside Postgres. The service waits for Postgres to become healthy before starting, and has a healthcheck of its own against `/health`. `.env` points `PG_HOST` at `127.0.0.1` for host-side tooling; compose overrides it to the `postgres` service name inside the network.

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
make full-lint        # format + lint — run before committing
make check-format     # non-rewriting format check (what CI runs)
make vulncheck        # scan dependencies for known vulnerabilities
make generate-mocks   # regenerate mocks
```

### Tests

Repository tests run against a real Postgres. `internal/utils/testdb` applies the
migrations from `db/migrations` and truncates every table before each test, so
the test schema is never a second copy of the DDL.

```sh
make test-db-up        # start just Postgres
make test-integration  # run the suite against it
make test-db-down      # stop it
make test-db-reset     # wipe the data directory and start over
```

The database files live in `infra/postgres/` (a bind mount, ignored by git), so
you can see and delete them like any other file. Because it is not a named
volume, `docker compose down -v` will not reset the database — use
`make test-db-reset`.

`make test` alone also works: tests needing a database skip themselves when none
is reachable. `make test-integration` fails instead of skipping, which is what CI
runs so a missing database can never pass silently.

CI (`.github/workflows/checks.yml`) runs format, lint, vulncheck, test and a Docker image build on every PR.

## Docs

- [Project layout conventions](./docs/structure.md)
- [Code style guide](./docs/code_style.md)
