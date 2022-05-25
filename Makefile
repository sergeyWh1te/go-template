tools:
	cd tools && go mod tidy && go mod vendor && go mod verify && go generate -tags tools
.PHONY: tools	

vendor:
	go mod tidy && go mod vendor && go mod verify
.PHONY: vendor	

build:
	go build -o ./bin/service ./cmd/service
.PHONY: build	

fmt:
	go vet ./cmd/... && fmt ./internal/...

vet:
	go vet ./cmd/... && fmt ./internal/...

imports:
	bin/goimports -local github.com/lidofinance/go-template -w -d $(shell find . -type f -name '*.go'| grep -v "/vendor/\|/.git/\|/tools/")

lint:
	bin/golangci-lint run --config=.golangci.yml --fix ./..

full-lint: imports fmt vet lint	
.PHONY: full-lint	
