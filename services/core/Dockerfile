FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

COPY pkg ./pkg

COPY services/core/go.mod services/core/go.sum ./services/core/

WORKDIR /app/services/core

RUN go mod download

COPY services/core/ .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service ./cmd/service

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/services/core/service .

COPY --from=builder /app/services/core/internal/database/migrations ./internal/database/migrations

EXPOSE 8082

CMD ["./service"]