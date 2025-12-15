FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o service ./cmd/service

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/service .

COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

EXPOSE 8083

CMD ["./service"]
