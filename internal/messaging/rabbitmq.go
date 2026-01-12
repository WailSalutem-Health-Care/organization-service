package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName = "wailsalutem.events"
	ExchangeType = "topic"
)

// Publisher handles publishing events to RabbitMQ
type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

// NewPublisher creates a new RabbitMQ publisher
func NewPublisher() (*Publisher, error) {
	rabbitmqURL := os.Getenv("RABBITMQ_URL")
	if rabbitmqURL == "" {
		rabbitmqURL = "amqp://admin:admin123@localhost:5672/"
	}

	log.Printf("Connecting to RabbitMQ at: %s", maskPassword(rabbitmqURL))

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange (topic exchange for flexible routing)
	err = channel.ExchangeDeclare(
		ExchangeName, // name
		ExchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	log.Printf("✓ Connected to RabbitMQ and declared exchange: %s", ExchangeName)

	return &Publisher{
		conn:     conn,
		channel:  channel,
		exchange: ExchangeName,
	}, nil
}

// Publish publishes an event to RabbitMQ with the specified routing key
func (p *Publisher) Publish(ctx context.Context, routingKey string, eventData interface{}) error {
	if p == nil || p.channel == nil {
		log.Printf("Warning: RabbitMQ publisher not initialized, skipping event: %s", routingKey)
		return nil
	}

	body, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		p.exchange, // exchange
		routingKey, // routing key (e.g., "patient.created")
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,  // persist messages to disk
			Timestamp:    time.Now().UTC(), // Explicitly set to UTC
			MessageId:    fmt.Sprintf("%d", time.Now().UnixNano()),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish event to %s: %w", routingKey, err)
	}

	log.Printf("Published event: %s", routingKey)
	return nil
}

// Close closes the RabbitMQ connection
func (p *Publisher) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			log.Printf("Error closing RabbitMQ channel: %v", err)
		}
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// maskPassword masks the password in RabbitMQ URL for logging
func maskPassword(url string) string {
	// Simple masking for security in logs
	return "amqp://***:***@..."
}
