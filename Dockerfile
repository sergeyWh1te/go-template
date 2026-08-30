# Build stage
FROM golang:1.27.0-alpine3.24 AS builder

WORKDIR /go/src/app

# Passed in by the build (see the docker-build Makefile target). The .git
# directory is excluded from the build context, so the commit cannot be read here.
ARG COMMIT=unknown

# Download modules first so the layer is reused whenever only sources change.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 keeps the binary static so it runs on a bare alpine image.
# The commit is stamped into the build_info metric.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/sergeyWh1te/go-template/internal/connectors/metrics.Commit=${COMMIT}" \
    -o ./bin/service ./cmd/service

# Run stage
FROM alpine:3.24

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/app/bin/service ./service
COPY --from=builder /go/src/app/db/migrations ./db/migrations

EXPOSE 8080

USER nobody

CMD ["./service"]
