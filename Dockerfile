# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /go/src/app

COPY . .

RUN go build -o ./bin/main ./cmd/service

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/app/bin ./bin
COPY --from=builder /go/src/app/db/migrations ./db/migrations

EXPOSE 8080

CMD ["./bin/main"]
