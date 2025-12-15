FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

# Копируем корневой go.mod для монорепо
COPY go.mod ./
COPY go.sum* ./

# Копируем общие пакеты из корня
COPY pkg/ ./pkg/

# Копируем код сервиса analyzer
COPY services/analyzer/ ./services/analyzer/

# Переходим в директорию сервиса
WORKDIR /app/services/analyzer

# Скачиваем зависимости
RUN go mod download

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/service \
    ./cmd/service/main.go

FROM alpine:latest AS runtime



RUN addgroup -g 1000 appuser && adduser -D -u 1000 -G appuser appuser
WORKDIR /app

COPY --from=builder /app/bin/service /app/service

RUN chown -R appuser:appuser /app
USER appuser

FROM runtime AS service
EXPOSE 8080
ENTRYPOINT ["/app/service"]
