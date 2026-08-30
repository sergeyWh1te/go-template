## 30.08.2026

### Connectors

1. `postgres.Connect` no longer caches a failed connection — `sync.Once` kept the first failure forever and returned a broken handle with a `nil` error on every later call. Memoized behind a `sync.Mutex`, storing the handle only after a successful ping.
2. `postgres`: a handle that fails its ping is closed and dropped, so the next call dials again instead of reusing a connection that never worked. The ping is bounded by a 5s context.
3. `postgres`: `MaxIdleConns` was 60 against `MaxOpenConns` 25 — `database/sql` silently clamps idle to open, so the constant lied. Lowered to 10, and added `ConnMaxLifetime`/`ConnMaxIdleTime`. Pool settings are applied before the ping, not after.
4. Added `postgres.Close()`. `Connect` memoizes the handle in a package-level var, and `db.Close()` alone did not clear it — a later `Connect` would have returned the closed handle with a nil error, the same class of bug as the cached failure. Close now closes *and* clears, and is safe to call twice.
5. Dropped `simple_protocol=true` from the pgx config. It disables prepared statements and is only needed behind a transaction-mode pooler such as pgbouncer; without one it costs performance and gives up protocol-level protection against injection through parameters.
6. `metrics`: `promauto.NewCounter` published `build_info` to the *global default* registry while `Store` held its own — the metric never appeared on `/metrics`. Now registered with `promauto.With(promRegistry)`.
7. `metrics`: registered the Go and process collectors, and added `commit` (stamped via `-ldflags`) and `version` labels to `build_info`. Metric name lowercased to `<app>_metric_build_info`.
8. `routes.go`: `/metrics` serves the app's registry via `promhttp.HandlerFor` instead of `promhttp.Handler()`, which exposed the unused default registry.
9. `logger.New` no longer returns an error: an unset or unrecognized `LOG_LEVEL` falls back to info rather than refusing to start. `LOG_FORMAT` matching is case-insensitive.
10. `env`: replaced a bare type assertion on the viper error with `errors.AsType`, and fixed the missing-file check — `SetConfigFile` makes viper report a missing file as `*fs.PathError`, not `ConfigFileNotFoundError`, so startup would have failed instead of falling through to the shell environment. Added `READ_ENV_FROM_SHELL` (as `onchain-mon` does) for running with no `.env` at all.

### Graceful shutdown

It was graceful in name only — three separate bugs, each reproduced before being fixed.

11. `RunHTTPServer` passed the already-canceled context to `server.Shutdown`. `<-ctx.Done()` unblocks *because* the context is canceled, and Shutdown returns immediately on a canceled context — so no in-flight request was ever drained and the process could exit mid-request. Shutdown now gets its own `context.WithoutCancel` plus a 15s budget, and gives up on that deadline rather than hanging on a stuck handler.
12. `ListenAndServe` returned `http.ErrServerClosed` into the errgroup, which keeps only the *first* non-nil error. On a normal stop that was always ErrServerClosed, which `main` then filtered out — so a real shutdown failure could never be logged. ErrServerClosed is absorbed in the goroutine that produced it, and `main` logs whatever survives.
13. `main` relied on `defer stop()`, which runs after the deferred database close, so a second Ctrl-C during cleanup was swallowed. Signal trapping is released right after `g.Wait()`.
14. Shutdown order is now explicit in `main`: HTTP drains, then `stop()`, then `postgres.Close()`. The database goes last so handlers still finishing keep their connections.
15. `docker-compose.yml`: added `stop_grace_period: 30s`. Docker's default is 10s — *shorter* than the app's own 15s drain budget — so Docker would have SIGKILLed the container mid-drain and the graceful shutdown could never complete under load.

### Tests

16. Added `internal/utils/testdb` — `testdb.New(t)` connects, applies `db/migrations` via golang-migrate as a *library*, truncates every table with `restart identity`, and registers cleanup. The migrations are the single source of truth: no second copy of the DDL to drift.
17. Skip/fail split: no database and no `TEST_PG_DSN` → the test skips, so `go test ./...` passes on a machine with nothing running; `TEST_PG_DSN` set but unusable → the test fails, so a broken CI service container cannot look like a green run.
18. Added the first test in the repo: `internal/pkg/users/repository/repository_test.go` covers Create, Get, the missing-row case, and the per-test truncation itself — doubling as the worked example of `testdb`. The shutdown and connector fixes above were each verified by a test during development (the drain test was confirmed to fail against the previous implementation with `context canceled`), but those tests were dropped afterwards to keep the template lean.
19. `TEST_PG_DSN` is `unexport`ed in the Makefile — the blanket `export` would otherwise make every plain `make test` look like a configured database and turn the skips into failures.

### Tooling

20. Makefile: `include .env` → `-include .env`, so `make tools` works on a fresh checkout before `.env` exists.
21. Makefile: tool versions pinned as variables; added `check-format`, `vulncheck`, `test`, `test-integration`, `test-race`, `cover`, `tidy`, `generate-mocks`, `docker-build`, `logs`, `test-db-up`, `test-db-down` and `test-db-reset`; `fmt` now runs `golangci-lint fmt`; `docker-compose` → `docker compose`; removed `vendor` (the tree has no `vendor/`).
22. `make build` stamps the git commit into `metrics.Commit`.
23. `.golangci.yml`: added `errorlint`, `modernize`, `perfsprint`, `copyloopvar`, `gochecknoinits`, `usestdlibvars`, `usetesting`, `mnd`; `govet` switched to `enable-all` minus `fieldalignment`; added exclusion presets and a mocks exclusion for formatters. All findings fixed (`errors.AsType`, `WaitGroup.Go`, named constants in the concurrency examples).
24. Pinned golangci-lint v2.13.2 and govulncheck v1.7.0 — the previous pins were built with go1.26 and refuse to load against the `go 1.27` directive.
25. Added `golang-migrate/migrate/v4` as a direct dependency (previously only the `bin/migrate` CLI).

### Docker

26. Dockerfile: `golang:1.23-alpine` could not build a `go 1.27` module — now `golang:1.27.0-alpine3.24` on `alpine:3.24`. Added module-download layer caching, `CGO_ENABLED=0`, `-s -w`, a `COMMIT` build-arg and `USER nobody`. Only the service binary is copied out, not the whole `bin/`.
27. Added `.dockerignore` (previously absent: the whole tree, including `.git` and `bin`, was sent as build context). `infra` is excluded too, or the local database would be uploaded with it.
28. `docker-compose.yml`: the service could never reach Postgres — `.env` sets `PG_HOST=127.0.0.1`, which inside the container is the service itself. Overridden to the `postgres` service name. Added a `/health` healthcheck, log rotation, `restart: unless-stopped` and env-driven credentials/ports.
29. Postgres data moved from the `postgres_data` named volume to a bind mount at `infra/postgres/`, so the local database is visible in the working tree. `infra/postgres/.gitignore` (`*` plus `!.gitignore`) keeps the mount point tracked while the data stays out of git. `PGDATA` still points at the `pgdata` subdirectory: initdb refuses a data directory it does not own, so Postgres creates it inside the mount itself. Since `-v` only removes named volumes, `make test-db-reset` is what clears the data now.

### CI

30. Added `.github/workflows/checks.yml`: `format`, `lint`, `vulncheck`, `test` and `docker`. `test` is gated behind `vulncheck` and runs `make test-integration` against a Postgres service container, so the database-backed tests can never silently skip.
31. Bumped the workflow actions to majors that run on Node 24, silencing the runner's "Node 20 is being deprecated" warning: `actions/checkout` v4 → v7.0.1, `actions/setup-go` v5 → v7.0.0, `docker/setup-buildx-action` v3 → v4.3.0, `docker/build-push-action` v6 → v7.3.0.
32. Documented the "Failed to save: Unable to reserve cache" line in the workflow header. It is not a failure: the parallel jobs share one cache key derived from `go.sum` and race to save it, so every loser logs it while one job stores the entry and the rest reuse it next run. Left `cache: true` on all jobs rather than adding a read/write split, which would duplicate `setup-go`'s own key handling for no real gain.

### Dependencies

33. Security: `golang.org/x/text` v0.29.0 → v0.39.0 (GO-2026-5970, reachable from this code) and `golang.org/x/sys` v0.35.0 → v0.44.0 (GO-2026-5024).
34. Updated every remaining outdated direct dependency, leaving `make outdated` empty: `go-chi/chi/v5` v5.3.0 → v5.3.2, `jackc/pgx/v5` v5.9.2 → v5.10.0, `prometheus/client_golang` v1.23.2 → v1.24.1, `stretchr/testify` v1.11.1 → v1.12.1, `golang.org/x/sync` v0.21.0 → v0.22.0. Transitively `prometheus/common` v0.70.1, `prometheus/procfs` v0.21.1, `golang.org/x/sys` v0.47.0, `golang.org/x/text` v0.40.0, `go.yaml.in/yaml/v3` v3.0.5.


### Housekeeping

35. Removed `internal/app/http_server/shutdown_test.go` and `internal/connectors/postgres/postgres_test.go`. They served their purpose during development — each bug above was reproduced before being fixed — but a template does not need to carry the regression suite for them. `repository_test.go` stays as the one worked example of `testdb`. Side effect: the suite no longer spends 15s waiting out the drain-budget test.

### Documentation

36. Rewrote `docs/structure.md`. It described a hypothetical `my-awesome-go-project` with `api/`, `assets/`, `web/`, `website/` and `vendor/` — none of which exist here — while the real `infra/` and `docs/` were missing from the tree. The actual layout comes first now, annotated folder by folder and verified against the working tree, followed by a request-flow diagram and a step-by-step "Adding a domain" section; the golang-standards layout stays below as a reference. Fixed the broken tree glyphs (ASCII `|` mixed with `│`) and translated the leftover Russian.
37. Added `CLAUDE.md` and kept it in step with the code.

**Verification.** Integration tests pass against a real database; the rebuilt container comes up healthy; `/example` returns `{"ID":1}` through the full handler → usecase → repository → Postgres path; `/metrics` carries `build_info` plus the Go and process collectors; `docker compose stop` exits 0, not 137. Lint, race and govulncheck are clean.

## 29.05.2026 (2)

1. Replaced `logrus` with stdlib `log/slog` — removed `github.com/sirupsen/logrus` and `github.com/evalphobia/logrus_sentry` dependencies entirely.
2. Rewrote `internal/connectors/logger/logger.go` — returns `*slog.Logger` with `TextHandler` or `JSONHandler` based on `LOG_FORMAT`; log level parsed via `slog.Level.UnmarshalText`.
3. Updated `internal/utils/deps/logger.go` — `Logger` interface now uses slog-style structured signatures (`msg string, args ...any`).
4. Updated all call sites to structured key-value logging (`"err", err` instead of `fmt.Errorf` wrapping).

## 29.05.2026

1. Migrated `.golangci.yml` to golangci-lint v2 format — added `version: "2"`, moved formatters (`gofmt`, `goimports`) to `formatters` section, updated linter names (`gomnd` → `mnd`), removed deprecated linters (`exportloopref`, `stylecheck`, `typecheck`), migrated `issues.*` to `linters.exclusions.*`.
2. Fixed critical bug in `internal/env/env.go` — config read errors were silently swallowed due to checking wrong variable.
3. Fixed `MaxIdleConns` in `internal/connectors/postgres/postgres.go` — was `60 * int(time.Second)` (60 billion), now `60`.
4. Fixed broken regex in `internal/connectors/metrics/prometheus.go` — JavaScript-style `/-|\ /g` replaced with valid Go `[-\s]+`.
5. Fixed HTTP response ordering in `user_example` handler — headers are now set after error checks; errors return proper 4xx/5xx status codes.
6. Fixed race condition in `cmd/shared_memory` — `s.m.RLocker().Lock()` replaced with `s.m.RLock/RUnlock()`.
7. Fixed `cmd/fan_out` — was a load balancer (`select` to either channel), now true fan-out (sends to both channels).
8. Fixed `cmd/worker` — `fmt.Sprint` (result discarded) replaced with `fmt.Println`.
9. Fixed invalid SQL in `users` repository — `insert into users ()` → `insert into users default values`.
10. Removed deprecated `middleware.RealIP` from HTTP server (security vulnerability).
11. Removed `cmd/migrator` binary — replaced by `./bin/migrate` CLI invoked directly from Makefile targets.
12. Removed `tools/` Go module — replaced by explicit `go install pkg@version` calls in `make tools`.
13. Replaced hardcoded `POSTGRESQL_URL` in Makefile — now assembled from `PG_*` vars loaded from `.env`.
14. Added full set of migration targets to Makefile: `migrate`, `migrate-down`, `migrate-step-down`, `migrate-version`, `migrate-force`, `migrate-create`, `migrate-drop`.
15. Added `make outdated` target using `go-mod-outdated` to list available dependency updates.
16. Added `govulncheck` to `make tools`. (No Makefile target came with it at the time; `make vulncheck` was added on 30.08.2026.)
17. Fixed Makefile: added missing `.PHONY` declarations, removed duplicate `full-lint` target, fixed `./cmd...` → `./cmd/...`.
18. Fixed Docker: `golang:1.26.3-alpine` (nonexistent) → `golang:1.23-alpine`; removed Russian comment.
19. Updated `docker-compose.yml` to Compose v2 format — removed deprecated `version` field, upgraded Postgres 14 → 17, added named volume for data persistence, renamed network, removed deprecated `links`.
20. Updated `README.md` — rewritten as a self-presenting project page with dependency table, quick start, Docker instructions, project structure tree, and migration reference.
21. Added `CLAUDE.md` — guidance file for Claude Code with commands, architecture overview, and code style rules.

## 30.06.2022

1. Provided example how to use logger.
2. Added `docs/structure` — how to structure a Go app.
3. Added `docs/code_style` — how to write a Go app.
