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
16. Added `make govulncheck` — vulnerability scanning via `govulncheck`.
17. Fixed Makefile: added missing `.PHONY` declarations, removed duplicate `full-lint` target, fixed `./cmd...` → `./cmd/...`.
18. Fixed Docker: `golang:1.26.3-alpine` (nonexistent) → `golang:1.23-alpine`; removed Russian comment.
19. Updated `docker-compose.yml` to Compose v2 format — removed deprecated `version` field, upgraded Postgres 14 → 17, added named volume for data persistence, renamed network, removed deprecated `links`.
20. Updated `README.md` — rewritten as a self-presenting project page with dependency table, quick start, Docker instructions, project structure tree, and migration reference.
21. Added `CLAUDE.md` — guidance file for Claude Code with commands, architecture overview, and code style rules.

## 30.06.2022

1. Provided example how to use logger.
2. Added `docs/structure` — how to structure a Go app.
3. Added `docs/code_style` — how to write a Go app.
