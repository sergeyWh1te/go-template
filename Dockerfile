# Build stage
FROM golang:1.20.5-alpine as builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /go/src/app

# Копируем файлы проекта
COPY . .

# Собираем приложение
RUN go build -o ./bin/main ./cmd/service
RUN go build -o ./bin/migrator ./cmd/migrator

RUN echo $PG_HOST

# Run stage
FROM alpine:latest

# Устанавливаем необходимые пакеты
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /root/

# Копируем скомпилированные бинарники из предыдущего этапа
COPY --from=builder /go/src/app/bin ./bin
COPY --from=builder /go/src/app/db/migrations ./db/migrations

# Открываем порт 8080
EXPOSE 8080

CMD ["./bin/main"]