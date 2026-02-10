package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client wraps RabbitMQ connection and channel
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

// Message represents a queue message
type Message struct {
	Type      string                 `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	Payload   map[string]interface{} `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
}

// Connect establishes connection to RabbitMQ
func Connect(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: channel,
	}, nil
}

// DeclareQueue ensures a queue exists
func (c *Client) DeclareQueue(queueName string) error {
	_, err := c.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

// Publish sends a message to a queue
func (c *Client) Publish(ctx context.Context, queueName string, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Consume starts consuming messages from a queue
func (c *Client) Consume(queueName string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		queueName,
		"",    // consumer tag
		false, // auto-ack (we'll ack manually)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

// Close closes the connection
func (c *Client) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	return c.conn.Close()
}
