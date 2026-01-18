package messaging

import "context"

// PublisherInterface defines the contract for event publishing
// This allows for easy mocking in tests
type PublisherInterface interface {
	Publish(ctx context.Context, routingKey string, eventData interface{}) error
	Close() error
}

// Ensure Publisher implements PublisherInterface
var _ PublisherInterface = (*Publisher)(nil)
