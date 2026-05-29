#!make
include .env
export

POSTGRESQL_URL := postgres://$(PG_USERNAME):$(PG_PASSWORD)@$(PG_HOST):$(PG_PORT)/$(PG_DATABASE)?sslmode=$(PG_SSLMODE)

### GO tools
.PHONY: tools
tools:
	@echo "Installing dev tools into ./bin/ ..."
	GOBIN=$(PWD)/bin go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	GOBIN=$(PWD)/bin go install github.com/vektra/mockery/v3@v3.7.0
	GOBIN=$(PWD)/bin go install golang.org/x/tools/cmd/goimports@v0.45.0
	GOBIN=$(PWD)/bin go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1
	GOBIN=$(PWD)/bin go install golang.org/x/vuln/cmd/govulncheck@v1.3.0
	GOBIN=$(PWD)/bin go install github.com/psampaz/go-mod-outdated@v0.9.0

.PHONY: outdated
outdated:
	go list -u -m -json all | bin/go-mod-outdated -update -direct

.PHONY: vendor
vendor:
	go mod tidy && go mod vendor && go mod verify

build:
	go build -o ./bin/service ./cmd/service
.PHONY: build

.PHONY: fmt
fmt:
	go fmt ./cmd/... && go fmt ./internal/...

.PHONY: vet
vet:
	go vet ./cmd/... && go vet ./internal/...

.PHONY: imports
imports:
	bin/goimports -local github.com/sergeyWh1te/go-template -w -d $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/\|/tools/")

.PHONY: lint
lint:
	bin/golangci-lint run --config=.golangci.yml --fix ./cmd/... ./internal/...

.PHONY: full-lint
full-lint: imports fmt vet lint

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

.PHONY: up
up:
	UID_GID="$$(id -u):$$(id -g)" docker-compose up -d

.PHONY: up-rebuild
up-rebuild:
	UID_GID="$$(id -u):$$(id -g)" docker-compose up -d --build

.PHONY: down
down:
	UID_GID="$$(id -u):$$(id -g)" docker-compose down

.PHONY: run
run:
	$(MAKE) up && $(MAKE) migrate