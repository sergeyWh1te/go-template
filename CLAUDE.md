# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

**Setup (first time):**
```sh
make tools           # go install dev tools into ./bin/ (golangci-lint, mockery, goimports, migrate, govulncheck, go-mod-outdated)
cp sample.env .env   # the Makefile does `-include .env`; migration targets need the PG_* vars
make up              # docker compose up -d (postgres + service)
make migrate         # apply migrations
```

**Build and run:**
```sh
make build      # outputs ./bin/service, stamping the git commit into the build_info metric
./bin/service   # run the HTTP server (listens on $PORT; sample.env says 8080)
make run        # make up + make migrate
make up / down / logs / up-rebuild    # docker compose lifecycle
make docker-build                     # build the image with the commit build-arg
```

**Linting and formatting:**
```sh
make full-lint     # format + lint — use this before committing
make format        # imports + fmt + vet
make fmt           # golangci-lint fmt (gofmt + goimports)
make check-format  # same, but --diff and non-rewriting (what CI runs)
make lint          # golangci-lint run
make fix-lint      # golangci-lint run --fix
make vulncheck     # govulncheck -show verbose ./...
make outdated      # list dependencies with available updates
```

**Tests:**
```sh
make test-db-up        # start just Postgres (fixtures need a real database)
make test-integration  # run everything against it — fails if the DB is missing
make test              # same tests, but DB-backed ones skip when none is running
make test-race         # with -race
make cover             # coverage profile + total
make test-db-down      # stop it
make test-db-reset     # wipe the data directory and start fresh
```

Postgres data is a **bind mount** at `infra/postgres/`, not a named volume, so the
local database is visible in the working tree. Consequences: `docker compose down -v`
does **not** clear it (`-v` only removes named volumes) — use `make test-db-reset`;
`PGDATA` points at the `pgdata` subdirectory so Postgres creates it with the
ownership initdb requires; and `infra/postgres/.gitignore` ignores everything but
itself, so the mount point exists on a fresh clone while the data never enters git.
`infra` is in `.dockerignore` — otherwise the database would be sent as build context.

Single package / single test: `go test ./internal/pkg/users/...`, `go test -run TestName ./internal/...`.

Existing test: `internal/pkg/users/repository/repository_test.go` — the domain
layer against a real database, and the worked example of how to use `testdb`.

**Database migrations** (all target `db/migrations`, URL assembled from `PG_*` in `.env`):
```sh
make migrate                    # apply all pending migrations
make migrate-down               # roll back all migrations
make migrate-step-down          # roll back one migration
make migrate-version            # print current migration version
make migrate-force v=<version>  # force-set version (recover from dirty state)
make migrate-create name=<name> # create a new sequential up/down migration pair
make migrate-drop               # drop everything in the database
```

**Generate mocks:**
```sh
make generate-mocks   # go generate ./internal/...
```

## Architecture

Layered Go service template. The request flow is:

```
cmd/service/main.go
  → internal/app/http_server/   (composition root — wires everything)
    → internal/http/handlers/   (HTTP handlers)
      → internal/pkg/<domain>/usecase/     (business logic)
        → internal/pkg/<domain>/repository/ (DB access)
```

### Key structural conventions

**`internal/app/http_server/`** — the composition root, package name `server`. `server.go` defines `App` and `RunHTTPServer`; `routes.go` registers all routes and middleware; `repository.go` and `usecase.go` are factory files that instantiate concrete implementations behind their interfaces. When adding a domain, wire it in all three.

**`internal/pkg/<domain>/`** — one folder per domain (`users` is the reference implementation):
- `usecase.go` — `Usecase` interface, with a `//go:generate ./../../../bin/mockery --name Usecase` directive
- `repository.go` — `Repository` interface, same directive pattern
- `entity/` — domain structs (with `db:` tags for sqlx)
- `usecase/usecase.go` — concrete implementation; `New` takes the `Repository` interface
- `repository/repository.go` — concrete implementation over `*sqlx.DB`
- `mocks/` — auto-generated, never edit by hand (also excluded from linting)

**`internal/http/handlers/<name>/`** — one folder per endpoint. Handlers depend on the domain `Usecase` interface and the `deps.Logger` interface, never on concrete structs. See `routes.go:24` — `userexample.New(a.Logger, a.usecase.User)`.

**`internal/connectors/`** — infrastructure adapters: `postgres` (pgx stdlib driver wrapped in sqlx), `logger` (stdlib `log/slog`), `metrics` (Prometheus registry + build-info counter). The pgx connection deliberately uses the **extended** protocol: `simple_protocol=true` disables prepared statements and is only warranted behind a transaction-mode pooler such as pgbouncer.

Two rules hold across the connectors, and both exist because breaking them hides outages:
- **Never cache a failed connection.** Shared clients are memoized behind a `sync.Mutex`, not `sync.Once` — a `Once` keeps the first failure forever and hands back a broken client with a `nil` error on every later call. `postgres.Connect` stores the handle only after the ping succeeds, and closes it otherwise so the next call redials.
- **Register metrics into the app registry**, via `promauto.With(promRegistry)`. Bare `promauto.New*` publishes to the global default registry, which `/metrics` does not serve — the metric silently never appears. `routes.go` serves `promhttp.HandlerFor(a.Metrics.Prometheus, ...)` for the same reason.

`logger.New` returns just a `*slog.Logger` (no error): an unrecognized `LOG_LEVEL` falls back to info rather than refusing to start. `metrics.Commit` is stamped at build time via `-ldflags` by `make build` and the `COMMIT` Docker build-arg.

**`internal/utils/deps/`** — narrow interfaces for injected dependencies. `deps.Logger` is the slog-shaped logging interface handlers and usecases accept.

**`internal/utils/testdb/`** — fixtures for tests that need a real database. `testdb.New(t)` connects, applies `db/migrations`, truncates every table with `restart identity`, and registers cleanup — so each test starts from an empty schema with predictable ids. It applies the real migrations rather than a copy of the DDL, so the test schema cannot drift from production.

Its skip/fail split is deliberate: with no database reachable and `TEST_PG_DSN` unset the test **skips**, so `go test ./...` passes on a laptop with nothing running; when `TEST_PG_DSN` *is* set but the database is unusable it **fails**, so a broken CI service container can never look like a green run. CI therefore runs `make test-integration` with `TEST_PG_DSN` set, never plain `make test`.

`TEST_PG_DSN` is deliberately `unexport`ed in the Makefile — the Makefile does a blanket `export`, and exporting it would make every plain `make test` look like a configured database, turning the skips into failures.

**`internal/env/`** — Viper-backed config read from `.env`, memoized with `sync.Once`. Every variable is listed in `sample.env`; add new ones there too.

`READ_ENV_FROM_SHELL=true` skips the file entirely and takes config from the environment — that is how the service runs in Docker and CI, where no `.env` exists. A missing `.env` is tolerated either way: `SetConfigFile` makes viper report it as `*fs.PathError` rather than `ConfigFileNotFoundError`, so both are checked.

**`cmd/`** — four independent binaries: `service` (the HTTP server, the real entry point), `worker` (errgroup daemon skeleton), `fan_out` and `shared_memory` (concurrency examples). Only `service` is built by `make build` and shipped in the Dockerfile.

### Logging

The project uses stdlib **`log/slog`** — there is no logrus and no Sentry integration. Log with structured key-value pairs: `log.Error("get user", "err", err)`, not formatted strings. `LOG_FORMAT=json` selects `JSONHandler`, anything else gives `TextHandler`; `LOG_LEVEL` is parsed by `slog.Level.UnmarshalText` (`debug`/`info`/`warn`/`error`).

A `depguard` rule in `.golangci.yml` still denies `github.com/sirupsen/logrus` outside `internal/connectors/logger`, `internal/utils/deps` and `internal/app` — treat it as a guard against reintroducing it.

### Concurrency model

`cmd/service/main.go` builds a signal-aware context (`signal.NotifyContext` on SIGINT/SIGTERM) and an `errgroup.Group`. Long-running goroutines are launched with `g.Go(...)`; the HTTP server registers a second goroutine that calls `Shutdown` on `ctx.Done()`. Follow this pattern for any new background work — and copy these three details with it, each of which was a real bug:

- **Never pass the triggering context to `Shutdown`.** `<-ctx.Done()` unblocks precisely because that context is canceled, and `Shutdown` returns immediately on a canceled context without draining anything. Derive a fresh one with `context.WithoutCancel` plus a timeout budget (`defaultShutdownTimeout`).
- **Absorb `http.ErrServerClosed` in the goroutine that produced it.** `errgroup` keeps only the first non-nil error, and on a clean stop `ListenAndServe` always wins that race — a caller that filters `ErrServerClosed` afterwards would discard the shutdown error instead.
- **Release the signal handler explicitly after `g.Wait()`**, not via `defer`. A deferred `stop()` runs after the deferred database close, so a second Ctrl-C during cleanup is swallowed rather than killing the process.

Shutdown order in `main` is: HTTP drains (`g.Wait()`) → `stop()` releases the signal handler → `postgres.Close()`. The database must go last, or handlers still finishing would lose their connections. `postgres.Close()` also clears the memoized handle, so a later `Connect` dials a fresh one instead of returning the closed one.

**`stop_grace_period` in docker-compose must exceed `defaultShutdownTimeout`.** Docker's default grace period is 10s against the app's 15s drain budget, so Docker would SIGKILL the container mid-drain; the compose file sets 30s. A clean stop shows exit code 0, not 137.

### Dependencies

Dependencies are resolved from the module cache — there is **no** `vendor/` directory and no `tools/` module. Tool versions are pinned as variables at the top of the Makefile and installed by `make tools` into `./bin/`. Use `make tidy` (`go mod tidy && go mod verify`) after changing dependencies, and `make outdated` to see what has newer releases.

## Code style rules (from `docs/code_style.md`)

- Panic only during initialization (`main`, connector setup). Never panic in handlers, usecases, or repositories.
- Accept interfaces, return concrete structs. Keep handler/usecase/repository structs private and export only `New`.
- Methods return `(value, error)` pairs. Never return a pointer to a reference type (`*[]T`, `*map[K]V`).
- Avoid `else` — use early returns.
- Compare strings against `""` for emptiness, and `len(x) == 0` for slices/maps.
- Do not pass HTTP DTOs across layers; convert to a domain struct in the handler via a constructor function.
- TODO comments must include a link to the tracking task.
- Test names use underscores, not spaces (the console replaces spaces on failure, making them harder to grep).

## Linting notes

`.golangci.yml` is golangci-lint **v2** format, kept in step with the `onchain-mon` config. Notable settings: line length 140 (`lll`), cyclomatic complexity 15, `funlen` 50 statements, `govet` with `enable-all` (minus `fieldalignment`), and a `gofmt` rewrite rule turning `interface{}` into `any`. Imports are grouped with the local prefix `github.com/sergeyWh1te/go-template` last. `mocks` is excluded from both linting and formatting.

Beyond the basics the config enables `errorlint` (no bare type assertions on errors), `modernize` (suggests current stdlib idioms — it is what pushed `errors.AsType` and `WaitGroup.Go` into this tree), `perfsprint`, `mnd`, `copyloopvar`, `gochecknoinits`, `usestdlibvars` and `usetesting`.

**Tool versions must be built with a Go at least as new as the `go` directive in `go.mod`.** go.mod targets **1.27**, so golangci-lint is pinned to v2.13.2 and govulncheck to v1.7.0; older builds refuse to load the config with *"the Go language version used to build golangci-lint is lower than the targeted Go version"*. Versions are pinned once as variables at the top of the Makefile and mirrored in `.github/workflows/checks.yml`.

## CI

`.github/workflows/checks.yml` runs five jobs on PRs and pushes to `main`: `format` (`make check-format`), `lint`, `vulncheck`, `test` and `docker`.

`test` is gated behind `vulncheck` (`needs:`) so runner time is not spent on a tree with known vulnerable dependencies. It brings up a Postgres service container and runs `make test-integration` with `TEST_PG_DSN` set — never plain `make test`, which would let the database-backed tests skip silently. `docker` builds the image and asserts the binary landed at `/app/service`, without pushing.

Tool versions in the workflow `env:` block must stay in step with the Makefile variables.
