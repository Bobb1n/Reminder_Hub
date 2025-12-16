package rabbitmq

import (
	"context"
	"fmt"
	"time"
	"reminder-hub/pkg/logger"

	"github.com/cenkalti/backoff/v4"
	"github.com/streadway/amqp"
)

type RabbitMQConfig struct {
	Host         string `env:"HOST" env-default:"localhost"`
	Port         int    `env:"PORT" env-default:"5672"`
	User         string `env:"USER" env-default:"rabbit"`
	Password     string `env:"PASSWORD" env-default:"rabbit"`
	ExchangeName string `env:"EXCHANGE" env-default:"donotmatter"`
	Kind         string `env:"KIND" env-default:"topic"`
}

// Initialize new channel for rabbitmq
func NewRabbitMQConn(cfg *RabbitMQConfig, ctx context.Context, log *logger.CurrentLogger) (*amqp.Connection, error) {
	connAddr := fmt.Sprintf(
		"amqp://%s:%s@%s:%d/",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second // Maximum time to retry
	maxRetries := 5                      // Number of retries (including the initial attempt)

	var conn *amqp.Connection
	var err error

	err = backoff.Retry(func() error {
		conn, err = amqp.Dial(connAddr)
		if err != nil {
			log.Error(ctx, "Failed to connect to RabbitMQ: %v. Connection information: %s", err, connAddr)
			return err
		}
		return nil
	}, backoff.WithMaxRetries(bo, uint64(maxRetries)))

	if err != nil {
		log.Error(ctx, "Failed to connect after %d retries", maxRetries)
		return nil, err
	}

	log.Info(ctx, "Connected to RabbitMQ")
	//savedCtx used to avoid panic, because the canceled ctx is invalid
	savedCtx := ctx
	go func() {
		<-ctx.Done()
		err := conn.Close()
		if err != nil {
			log.Error(savedCtx, "Failed to close RabbitMQ connection %v", err)
		}
		log.Info(savedCtx, "RabbitMQ connection is closed")
	}()

	return conn, err
}


