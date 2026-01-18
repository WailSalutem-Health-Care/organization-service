package testutil

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// PublishedEvent represents an event that was published to RabbitMQ
type PublishedEvent struct {
	RoutingKey string
	EventData  interface{}
	Timestamp  time.Time
	RawJSON    []byte
}

// MockPublisher is a mock implementation of RabbitMQ publisher for testing
// It stores all published events in memory and doesn't make any real RabbitMQ calls
type MockPublisher struct {
	mu     sync.RWMutex
	events []PublishedEvent
}

// NewMockPublisher creates a new mock RabbitMQ publisher
func NewMockPublisher() *MockPublisher {
	return &MockPublisher{
		events: make([]PublishedEvent, 0),
	}
}

// Publish stores an event in memory (no real RabbitMQ call)
func (m *MockPublisher) Publish(ctx context.Context, routingKey string, eventData interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Marshal to JSON to simulate real publishing
	jsonData, err := json.Marshal(eventData)
	if err != nil {
		return err
	}

	event := PublishedEvent{
		RoutingKey: routingKey,
		EventData:  eventData,
		Timestamp:  time.Now(),
		RawJSON:    jsonData,
	}

	m.events = append(m.events, event)
	return nil
}

// Close is a no-op for mock publisher
func (m *MockPublisher) Close() error {
	return nil
}

// Helper methods for test assertions

// GetAllEvents returns all published events
func (m *MockPublisher) GetAllEvents() []PublishedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy
	eventsCopy := make([]PublishedEvent, len(m.events))
	copy(eventsCopy, m.events)
	return eventsCopy
}

// GetEventsByKey returns all events with the specified routing key
func (m *MockPublisher) GetEventsByKey(routingKey string) []PublishedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []PublishedEvent
	for _, event := range m.events {
		if event.RoutingKey == routingKey {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetEventCount returns the total number of events published
func (m *MockPublisher) GetEventCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.events)
}

// GetEventCountByKey returns the number of events with the specified routing key
func (m *MockPublisher) GetEventCountByKey(routingKey string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, event := range m.events {
		if event.RoutingKey == routingKey {
			count++
		}
	}
	return count
}

// Reset clears all published events (for test cleanup)
func (m *MockPublisher) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.events = make([]PublishedEvent, 0)
}

// AssertEventPublished asserts that at least one event with the given routing key was published
func (m *MockPublisher) AssertEventPublished(t *testing.T, routingKey string) {
	t.Helper()

	count := m.GetEventCountByKey(routingKey)
	if count == 0 {
		t.Errorf("Expected event with routing key '%s' to be published, but found none", routingKey)
	}
}

// AssertEventNotPublished asserts that no events with the given routing key were published
func (m *MockPublisher) AssertEventNotPublished(t *testing.T, routingKey string) {
	t.Helper()

	count := m.GetEventCountByKey(routingKey)
	if count > 0 {
		t.Errorf("Expected no events with routing key '%s', but found %d", routingKey, count)
	}
}

// AssertEventCount asserts the exact number of events with the given routing key
func (m *MockPublisher) AssertEventCount(t *testing.T, routingKey string, expected int) {
	t.Helper()

	count := m.GetEventCountByKey(routingKey)
	if count != expected {
		t.Errorf("Expected %d events with routing key '%s', got %d", expected, routingKey, count)
	}
}

// GetLastEvent returns the most recently published event, or nil if no events
func (m *MockPublisher) GetLastEvent() *PublishedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.events) == 0 {
		return nil
	}

	// Return a copy of the last event
	lastEvent := m.events[len(m.events)-1]
	return &lastEvent
}

// GetLastEventByKey returns the most recently published event with the given routing key
func (m *MockPublisher) GetLastEventByKey(routingKey string) *PublishedEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.events) - 1; i >= 0; i-- {
		if m.events[i].RoutingKey == routingKey {
			event := m.events[i]
			return &event
		}
	}
	return nil
}
