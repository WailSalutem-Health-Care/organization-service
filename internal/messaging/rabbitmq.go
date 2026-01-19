package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	ExchangeName = "wailsalutem.events"
	ExchangeType = "topic"
)

var tracer = otel.Tracer("github.com/WailSalutem-Health-Care/organization-service/messaging")

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
// It includes OpenTelemetry tracing with context propagation
func (p *Publisher) Publish(ctx context.Context, routingKey string, eventData interface{}) error {
	// Start a span for the publish operation
	ctx, span := tracer.Start(ctx, "rabbitmq.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			attribute.String("messaging.system", "rabbitmq"),
			attribute.String("messaging.destination", p.exchange),
			attribute.String("messaging.routing_key", routingKey),
			attribute.String("messaging.protocol", "AMQP"),
		),
	)
	defer span.End()

	if p == nil || p.channel == nil {
		log.Printf("Warning: RabbitMQ publisher not initialized, skipping event: %s", routingKey)
		span.SetStatus(codes.Error, "publisher not initialized")
		return nil
	}

	body, err := json.Marshal(eventData)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to marshal event data")
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	span.SetAttributes(attribute.Int("messaging.message_payload_size_bytes", len(body)))

	// Prepare headers for trace context propagation
	headers := make(amqp.Table)
	
	// Inject trace context into message headers
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, &rabbitMQCarrier{headers: headers})

	err = p.channel.PublishWithContext(
		ctx,
		p.exchange, // exchange
		routingKey, // routing key (e.g., "patient.created")
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // persist messages to disk
			Timestamp:    time.Now(),
			MessageId:    fmt.Sprintf("%d", time.Now().UnixNano()),
			Headers:      headers, // Include trace context
		},
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to publish message")
		return fmt.Errorf("failed to publish event to %s: %w", routingKey, err)
	}

	span.SetStatus(codes.Ok, "message published successfully")
	log.Printf("Published event: %s", routingKey)
	return nil
}

// rabbitMQCarrier implements the TextMapCarrier interface for RabbitMQ headers
type rabbitMQCarrier struct {
	headers amqp.Table
}

// Get retrieves a value from the carrier
func (c *rabbitMQCarrier) Get(key string) string {
	if val, ok := c.headers[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// Set stores a key-value pair in the carrier
func (c *rabbitMQCarrier) Set(key, value string) {
	c.headers[key] = value
}

// Keys returns all keys in the carrier
func (c *rabbitMQCarrier) Keys() []string {
	keys := make([]string, 0, len(c.headers))
	for k := range c.headers {
		keys = append(keys, k)
	}
	return keys
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
