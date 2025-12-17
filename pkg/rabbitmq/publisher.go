package rabbitmq

import (
	"context"
	"reflect"
	"reminder-hub/pkg/logger"
	"sync"
	"time"

	"github.com/ettle/strcase"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

//go:generate mockery --name IPublisher
type IPublisher interface {
	PublishMessage(msg interface{}) error
	IsPublished(msg interface{}) bool
}

type Publisher struct {
	cfg               *RabbitMQConfig
	conn              *amqp.Connection
	log               logger.CurrentLogger
	ctx               context.Context
	publishedMessages map[string]bool
	mu                sync.Mutex
}

func (p *Publisher) PublishMessage(msg interface{}) error {

	data, err := jsoniter.Marshal(msg)

	if err != nil {
		p.log.Error(p.ctx, "Error in marshalling message to publish message %v", err)
		return err
	}
	//For different services it will be different typeNames:
	//- For CoreService it must be RawEmails
	//- For AnalyzerService it must be ParsedEmails
	typeName := reflect.TypeOf(msg).Elem().Name()
	snakeTypeName := strcase.ToSnake(typeName)

	p.log.Info(context.Background(), "SnakeTypeName of Publisher: ", snakeTypeName)

	channel, err := p.conn.Channel()
	if err != nil {
		p.log.Error(p.ctx, "Error in opening channel to consume message %v", err)
		return err
	}

	defer channel.Close()

	err = channel.ExchangeDeclare(
		snakeTypeName, // name
		p.cfg.Kind,    // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)

	if err != nil {
		p.log.Error(p.ctx, "Error in declaring exchange to publish message")
		return err
	}

	correlationId := ""

	if p.ctx.Value(echo.HeaderXCorrelationID) != nil {
		correlationId = p.ctx.Value(echo.HeaderXCorrelationID).(string)
	}

	publishingMsg := amqp.Publishing{
		Body:          data,
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		MessageId:     uuid.New().String(),
		Timestamp:     time.Now(),
		CorrelationId: correlationId,
	}

	err = channel.Publish(snakeTypeName, snakeTypeName, false, false, publishingMsg)

	if err != nil {
		p.log.Error(p.ctx, "Error in publishing message")
		return err
	}
	p.mu.Lock()
	p.publishedMessages[snakeTypeName] = true
	p.mu.Unlock()
	p.log.Info(p.ctx, "Published message: %s", publishingMsg.Body)
	return nil
}

func (p *Publisher) IsPublished(msg interface{}) bool {

	typeName := reflect.TypeOf(msg).Name()
	snakeTypeName := strcase.ToSnake(typeName)
	p.mu.Lock()
	isPublished := p.publishedMessages[snakeTypeName]
	p.mu.Unlock()

	return isPublished
}

func NewPublisher(ctx context.Context, cfg *RabbitMQConfig, conn *amqp.Connection, log *logger.CurrentLogger) IPublisher {
	return &Publisher{ctx: ctx, cfg: cfg, conn: conn, log: *log, publishedMessages: make(map[string]bool), mu: sync.Mutex{}}
}
