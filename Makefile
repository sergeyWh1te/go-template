#!make
# `-include` (not `include`): a missing .env must not break targets that do not
# need it, such as `make tools` on a fresh checkout.
-include .env
export

POSTGRESQL_URL := postgres://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE)?sslmode=$(PG_SSLMODE)

GOLANGCI_LINT_VERSION := v2.13.2
MOCKERY_VERSION       := v3.7.3
GOIMPORTS_VERSION     := v0.48.0
MIGRATE_VERSION       := v4.19.1
GOVULNCHECK_VERSION   := v1.7.0
GO_MOD_OUTDATED_VERSION := v0.9.0

### GO tools
.PHONY: tools
tools:
	@echo "Installing dev tools into ./bin ..."
	GOBIN=$(PWD)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	GOBIN=$(PWD)/bin go install github.com/vektra/mockery/v3@$(MOCKERY_VERSION)
	GOBIN=$(PWD)/bin go install golang.org/x/tools/cmd/goimports@$(GOIMPORTS_VERSION)
	GOBIN=$(PWD)/bin go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	GOBIN=$(PWD)/bin go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	GOBIN=$(PWD)/bin go install github.com/psampaz/go-mod-outdated@$(GO_MOD_OUTDATED_VERSION)

.PHONY: outdated
outdated:
	@echo "Checking for outdated modules..."
	go list -u -m -json -mod=mod all | ./bin/go-mod-outdated -update -direct

.PHONY: vulncheck
vulncheck:
	@echo "Running govulncheck..."
	./bin/govulncheck -show verbose ./...

.PHONY: tidy
tidy:
	go mod tidy && go mod verify

.PHONY: build
build:
	go build -ldflags="-X github.com/sergeyWh1te/go-template/internal/connectors/metrics.Commit=$(shell git rev-parse HEAD)" \
		-o ./bin/service ./cmd/service

### Formatting and linting
.PHONY: fmt
fmt:
	bin/golangci-lint fmt --config=.golangci.yml ./cmd/... ./internal/...

# Fails when anything is not formatted, without rewriting files (for CI).
.PHONY: check-format
check-format:
	bin/golangci-lint fmt --config=.golangci.yml --diff ./cmd/... ./internal/...

.PHONY: vet
vet:
	go vet ./cmd/... ./internal/...

.PHONY: imports
imports:
	bin/goimports -local github.com/sergeyWh1te/go-template -w $(shell find ./cmd ./internal -type f -name '*.go')

.PHONY: lint
lint:
	bin/golangci-lint run --config=.golangci.yml ./cmd/... ./internal/...

.PHONY: fix-lint
fix-lint:
	bin/golangci-lint run --config=.golangci.yml --fix ./cmd/... ./internal/...

.PHONY: format
format: imports fmt vet

# Kept as the pre-commit entry point.
.PHONY: full-lint
full-lint: format lint

### Tests
# Tests needing Postgres skip themselves when none is reachable, so `make test`
# stays green on a machine with nothing running. Bring one up with test-db-up to
# actually exercise them.
# unexport: the Makefile does `export`, and exporting this would make every
# `make test` run look like a configured database, turning skips into failures.
TEST_PG_DSN ?= postgres://postgres:postgres@127.0.0.1:5432/master?sslmode=disable
unexport TEST_PG_DSN

.PHONY: test
test:
	go test ./cmd/... ./internal/...

# Fails, rather than skips, if the database is missing — for CI and for checking
# the database-backed tests really ran.
.PHONY: test-integration
test-integration:
	TEST_PG_DSN="$(TEST_PG_DSN)" go test -count=1 ./cmd/... ./internal/...

.PHONY: test-race
test-race:
	go test -race ./cmd/... ./internal/...

.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./cmd/... ./internal/...
	go tool cover -func=coverage.out | tail -1

# Just the database, without the service — what the tests need locally.
.PHONY: test-db-up
test-db-up:
	docker compose up -d postgres
	@echo "waiting for postgres..."
	@until docker compose exec -T postgres pg_isready -U postgres -d master >/dev/null 2>&1; do sleep 1; done
	@echo "postgres is ready — run: make test-integration"

.PHONY: test-db-down
test-db-down:
	docker compose stop postgres

# The data lives in a bind mount, so `docker compose down -v` does not clear it
# (-v only removes named volumes). Removing the directory is the reset. It is
# written by the container's postgres user, hence docker rather than plain rm.
.PHONY: test-db-reset
test-db-reset:
	docker compose rm -sf postgres
	docker run --rm -v "$(PWD)/infra/postgres:/data" alpine:3.24 \
		sh -c 'rm -rf /data/pgdata'
	$(MAKE) test-db-up

### Mocks
.PHONY: generate-mocks
generate-mocks:
	go generate ./internal/...

### Migrations
.PHONY: migrate
migrate:
	bin/migrate -database ${POSTGRESQL_URL} -path db/migrations up

.PHONY: migrate-down
migrate-down:
	bin/migrate -database ${POSTGRESQL_URL} -path db/migrations down

.PHONY: migrate-step-down
migrate-step-down:
	bin/migrate -database ${POSTGRESQL_URL} -path db/migrations down 1

.PHONY: migrate-version
migrate-version:
	bin/migrate -database ${POSTGRESQL_URL} -path db/migrations version

.PHONY: migrate-force
migrate-force:
	bin/migrate -database ${POSTGRESQL_URL} -path db/migrations force $(v)

.PHONY: migrate-create
migrate-create:
	bin/migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: migrate-drop
migrate-drop:
	bin/migrate -database ${POSTGRESQL_URL} -path db/migrations drop -f

### Docker
.PHONY: docker-build
docker-build:
	docker build --build-arg COMMIT=$(shell git rev-parse HEAD) -t go-template:stable -f Dockerfile .

.PHONY: up
up:
	docker compose up -d

.PHONY: up-rebuild
up-rebuild:
	COMMIT=$(shell git rev-parse HEAD) docker compose up -d --build

.PHONY: down
down:
	docker compose down

.PHONY: logs
logs:
	docker compose logs -f --tail=100

.PHONY: run
run:
	$(MAKE) up && $(MAKE) migrate
