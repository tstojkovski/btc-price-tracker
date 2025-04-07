package store

import (
	"btc-price-tracker/internal/domain"
	"testing"
)

func TestMemoryStore_Store(t *testing.T) {
	// Create a store with capacity of 3
	store := NewMemoryStore(3)

	// Create test events
	event1 := domain.PriceUpdateEvent{Timestamp: 100, Price: 50000.0}
	event2 := domain.PriceUpdateEvent{Timestamp: 101, Price: 51000.0}
	event3 := domain.PriceUpdateEvent{Timestamp: 102, Price: 52000.0}
	event4 := domain.PriceUpdateEvent{Timestamp: 103, Price: 53000.0}

	// Store events
	store.Store(event1)
	store.Store(event2)
	store.Store(event3)

	// Verify latest event
	latest, exists := store.GetLatestEvent()
	if !exists {
		t.Fatal("Expected latest event to exist")
	}
	if latest.Timestamp != 102 {
		t.Errorf("Expected timestamp 102, got %d", latest.Timestamp)
	}

	// Test circular buffer behavior by adding a fourth event
	store.Store(event4)

	// The oldest event (event1) should be overwritten
	events := store.GetEventsSince(100)

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	// First event should now be event2
	if events[0].Timestamp != 101 {
		t.Errorf("Expected first event timestamp to be 101, got %d", events[0].Timestamp)
	}
}

func TestMemoryStore_GetEventsSince(t *testing.T) {
	store := NewMemoryStore(5)

	// Add events with different timestamps
	events := []domain.PriceUpdateEvent{
		{Timestamp: 100, Price: 50000.0},
		{Timestamp: 200, Price: 51000.0},
		{Timestamp: 300, Price: 52000.0},
		{Timestamp: 400, Price: 53000.0},
	}

	for _, e := range events {
		store.Store(e)
	}

	tests := []struct {
		name          string
		since         int64
		expectedCount int
	}{
		{"Get all events", 0, 4},
		{"Get events since 100", 100, 4},
		{"Get events since 101", 101, 3},
		{"Get events since 300", 300, 2},
		{"Get events since 500", 500, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := store.GetEventsSince(tc.since)
			if len(result) != tc.expectedCount {
				t.Errorf("Expected %d events, got %d", tc.expectedCount, len(result))
			}
		})
	}
}

func TestMemoryStore_GetLatestEvent(t *testing.T) {
	store := NewMemoryStore(3)

	// Test with empty store
	_, exists := store.GetLatestEvent()
	if exists {
		t.Error("Expected no event to exist in empty store")
	}

	// Add an event
	event := domain.PriceUpdateEvent{Timestamp: 100, Price: 50000.0}
	store.Store(event)

	// Get latest event
	latest, exists := store.GetLatestEvent()
	if !exists {
		t.Fatal("Expected latest event to exist")
	}
	if latest.Timestamp != 100 || latest.Price != 50000.0 {
		t.Errorf("Expected event {100, 50000.0}, got {%d, %.2f}", latest.Timestamp, latest.Price)
	}

	// Add another event
	event2 := domain.PriceUpdateEvent{Timestamp: 200, Price: 51000.0}
	store.Store(event2)

	// Get latest event again
	latest, exists = store.GetLatestEvent()
	if !exists {
		t.Fatal("Expected latest event to exist")
	}
	if latest.Timestamp != 200 || latest.Price != 51000.0 {
		t.Errorf("Expected event {200, 51000.0}, got {%d, %.2f}", latest.Timestamp, latest.Price)
	}
}
