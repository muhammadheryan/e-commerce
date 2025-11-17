package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	apiURL  string
	apiKey  string
}

func NewConsumer(host string, port int, user, password, apiURL, apiKey string) (*Consumer, error) {
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
		"order_expiration_exchange",
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp091.Table{"x-delayed-type": "direct"},
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Declare the queue
	_, err = channel.QueueDeclare(
		"order_expiration_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		"order_expiration_queue",
		"order_expiration",
		"order_expiration_exchange",
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: channel,
		apiURL:  apiURL,
		apiKey:  apiKey,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	// Set QoS to 1 - process one message at a time
	err := c.channel.Qos(1, 0, false)
	if err != nil {
		return err
	}

	msgs, err := c.channel.Consume(
		"order_expiration_queue",
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-msgs:
				if msg.DeliveryTag == 0 { // channel closed
					return
				}

				var orderMsg OrderExpirationMessage
				err := json.Unmarshal(msg.Body, &orderMsg)
				if err != nil {
					log.Printf("Failed to unmarshal message: %v", err)
					msg.Ack(false)
					continue
				}

				// Call cancel order API
				err = c.callCancelOrderAPI(orderMsg.OrderID, orderMsg.UserID)
				if err != nil {
					log.Printf("Failed to cancel order %d: %v", orderMsg.OrderID, err)
					// Negative ack to requeue
					msg.Nack(false, true)
					continue
				}

				// Success - acknowledge the message
				msg.Ack(false)
				log.Printf("Order %d cancelled successfully", orderMsg.OrderID)
			}
		}
	}()

	return nil
}

func (c *Consumer) callCancelOrderAPI(orderID, userID uint64) error {
	url := fmt.Sprintf("%s/internal/v1/order/%d/cancel", c.apiURL, orderID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	// Add authorization header using the API key (internal service key)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Service", "order-expiration-consumer")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 500 {
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}
