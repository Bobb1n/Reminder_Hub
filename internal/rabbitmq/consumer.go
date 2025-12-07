package rabbitmq

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
	"yfp/internal/logger"

	"github.com/ettle/strcase"
	"github.com/streadway/amqp"
)

type IConsumer[T any] interface {
	ConsumeMessage(msg interface{}, dependencies T) error
	IsConsumed(msg interface{}) bool
}

type Consumer[T any] struct {
	cfg              *RabbitMQConfig
	conn             *amqp.Connection
	log              *logger.CurrentLogger
	handler          func(queue string, msg amqp.Delivery, dependencies T) error
	ctx              context.Context
	consumedMessages map[string]bool
	mu               sync.Mutex
}

func (c *Consumer[T]) ConsumeMessage(msg interface{}, dependencies T) error {
	ch, err := c.conn.Channel()
	if err != nil {
		c.log.Error(c.ctx, "Error in opening channel to consume message", err)
		return err
	}
	//For different services it will be different typeNames:
	//- For AnalyzerService it must be RawEmails
	//- For For CollectorService it must be ParsedEmails
	typeName := reflect.TypeOf(msg).Name()
	snakeTypeName := strcase.ToSnake(typeName)

	err = ch.ExchangeDeclare(
		snakeTypeName, // name
		c.cfg.Kind,    // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)

	if err != nil {
		c.log.Error(c.ctx, "Error in declaring exchange to consume message %v", err)
		return err
	}

	q, err := ch.QueueDeclare(
		fmt.Sprintf("%s_%s", snakeTypeName, "queue"), // name (for ex: RawEmails -> raw_emails_queue)
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		c.log.Error(c.ctx, "Error in declaring queue to consume message")
		return err
	}

	err = ch.QueueBind(
		q.Name,        // queue name
		snakeTypeName, // routing key
		snakeTypeName, // exchange
		false,
		nil)
	if err != nil {
		c.log.Error(c.ctx, "Error in binding queue to consume message")
		return err
	}

	deliveries, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)

	if err != nil {
		c.log.Error(c.ctx, "Error in consuming message")
		return err
	}

	savedCtx := c.ctx
	go func() {
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			if err != nil {
				c.log.Error(savedCtx, "failed to close channel. Closed for queue: %s", q.Name)
			}
		}(ch)
		for {
			select {
			case <-c.ctx.Done():
				c.log.Info(savedCtx, "channel closed for for queue: %s", q.Name)
				return
			case delivery, ok := <-deliveries:
				if !ok {
					c.log.Error(savedCtx, "NOT OK deliveries channel closed for queue: %s", q.Name)
					ch.Close()
					return
				}

				//For ServiceAnalyzer it will be function related to LLM processing
				//For Collector it will be function, which stores values into DB
				err := c.handler(q.Name, delivery, dependencies)
				if err != nil {
					c.log.Error(savedCtx, err.Error())
					delivery.Nack(false, true)
				} else {
					c.mu.Lock()
					c.consumedMessages[snakeTypeName] = true
					c.mu.Unlock()
				}

				err = delivery.Ack(false)
				if err != nil {
					c.log.Error(savedCtx, "We didn't get a ack for delivery: %v", string(delivery.Body))
					delivery.Nack(false, true)
				}
			}
		}

	}()
	c.log.Info(c.ctx, "Waiting for messages in queue :%s. To exit press CTRL+C", q.Name)

	return nil
}

func (c *Consumer[T]) IsConsumed(msg interface{}) bool {
	timeOutTime := 20 * time.Second
	startTime := time.Now()
	timeOutExpired := false
	isConsumed := false

	for {
		if timeOutExpired {
			return false
		}
		if isConsumed {
			return true
		}

		time.Sleep(time.Second * 2)

		typeName := reflect.TypeOf(msg).Name()
		snakeTypeName := strcase.ToSnake(typeName)
		c.mu.Lock()
		_, isConsumed = c.consumedMessages[snakeTypeName]
		c.mu.Unlock()

		timeOutExpired = time.Since(startTime) > timeOutTime
	}
}

func NewConsumer[T any](ctx context.Context, cfg *RabbitMQConfig, conn *amqp.Connection, log *logger.CurrentLogger, handler func(queue string, msg amqp.Delivery, dependencies T) error) IConsumer[T] {
	return &Consumer[T]{
		ctx:              ctx,
		cfg:              cfg,
		conn:             conn,
		log:              log,
		handler:          handler,
		consumedMessages: make(map[string]bool),
		mu:               sync.Mutex{}}
}
