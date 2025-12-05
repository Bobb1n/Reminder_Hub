package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"
)

type Producer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type EmailMessage struct {
	EmailID       string `json:"email_id"`
	UserID        string `json:"user_id"`
	MessageID     string `json:"message_id"`
	FromAddress   string `json:"from_address"`
	Subject       string `json:"subject"`
	BodyText      string `json:"body_text"`
	DateReceived  string `json:"date_received"`
	SyncTimestamp string `json:"sync_timestamp"`
}

func NewProducer(url string) (*Producer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, errFailedToConnect(err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, errFailedToOpenChannel(err)
	}

	err = channel.ExchangeDeclare(
		"email.raw", // name
		"topic",     // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return nil, errFailedToDeclareExchange(err)
	}

	queue, err := channel.QueueDeclare(
		"email.raw.process", // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, errFailedToDeclareQueue(err)
	}

	err = channel.QueueBind(
		queue.Name,    // queue name
		"email.raw.*", // routing key pattern
		"email.raw",   // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, errFailedToBindQueue(err)
	}

	dlq, err := channel.QueueDeclare(
		"email.raw.dlq", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return nil, errFailedToDeclareDLQ(err)
	}

	err = channel.QueueBind(
		dlq.Name,        // queue name
		"email.raw.dlq", // routing key
		"email.raw",     // exchange
		false,
		nil,
	)
	if err != nil {
		return nil, errFailedToBindDLQ(err)
	}

	log.Info().Msg("RabbitMQ exchange, queues and bindings created successfully")

	return &Producer{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *Producer) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *Producer) PublishEmail(message *EmailMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return errFailedToMarshalMessage(err)
	}

	routingKey := "email.raw.process"

	err = p.channel.Publish(
		"email.raw", // exchange
		routingKey,  // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		log.Error().Err(err).Msgf("ERROR: Failed to publish message to RabbitMQ")
		return errFailedToPublishMessage(err)
	}

	log.Info().Msgf("Published email to RabbitMQ: %s for user %s (routing key: %s)",
		message.EmailID, message.UserID, routingKey)
	return nil
}
