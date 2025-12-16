FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Копируем корневой go.mod для монорепо
COPY go.mod ./
COPY go.sum* ./

# Копируем общие пакеты из корня
COPY pkg/ ./pkg/

# Копируем код сервиса auth
COPY services/auth/ ./services/auth/

# Переходим в директорию сервиса
WORKDIR /app/services/auth

# Скачиваем зависимости
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/service \
    ./cmd/service/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/migrate \
    ./cmd/migrate/main.go

FROM alpine:latest AS runtime

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app

COPY --from=builder /app/bin/service /app/service
COPY --from=builder /app/bin/migrate /app/migrate

COPY --from=builder /app/services/auth/migrations /app/migrations

RUN chown -R appuser:appuser /app
USER appuser

FROM runtime AS service
EXPOSE 8081
ENTRYPOINT ["/app/service"]

FROM runtime AS migrate
ENTRYPOINT ["/app/migrate"]
