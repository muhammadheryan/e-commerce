package rabbitmq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

type OrderExpirationMessage struct {
	OrderID   uint64    `json:"order_id"`
	UserID    uint64    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewPublisher(host string, port int, user, password string) (*Publisher, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%d/", user, password, host, port)
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declare the delayed exchange
	err = channel.ExchangeDeclare(
		"order_expiration_exchange", // name
		"x-delayed-message",         // type
		true,                        // durable
		false,                       // auto-delete
		false,                       // internal
		false,                       // no-wait
		amqp091.Table{"x-delayed-type": "direct"}, // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		"order_expiration_queue", // name
		true,                     // durable
		false,                    // auto-delete
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		"order_expiration_queue",    // queue name
		"order_expiration",          // routing key
		"order_expiration_exchange", // exchange
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Publisher{conn: conn, channel: channel}, nil
}

func (p *Publisher) PublishOrderExpiration(msg OrderExpirationMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	delayMs := int64((msg.ExpiresAt.Sub(time.Now()).Milliseconds()))
	if delayMs < 0 {
		delayMs = 0
	}

	return p.channel.Publish(
		"order_expiration_exchange", // exchange
		"order_expiration",          // routing key
		false,                       // mandatory
		false,                       // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp091.Table{
				"x-delay": delayMs,
			},
		},
	)
}

func (p *Publisher) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		p.conn.Close()
	}
	return nil
}
