FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o service ./cmd/service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/service .

EXPOSE 8080

CMD ["./service"]