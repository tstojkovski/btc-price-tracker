package service

import (
	"btc-price-tracker/internal/domain"
	"btc-price-tracker/internal/store"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBroadcastService_SubscribeUnsubscribe(t *testing.T) {
	memStore := store.NewMemoryStore(10)
	updateChan := make(chan domain.PriceUpdateEvent, 10)
	broadcastService := NewBroadcastService(memStore, updateChan)

	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	broadcastService.Start(ctx)

	// Subscribe a client
	clientChan := broadcastService.SubscribeClient()

	// Verify the client was added to the map
	if len(broadcastService.clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(broadcastService.clients))
	}

	// Unsubscribe the client
	broadcastService.UnsubscribeClient(clientChan)

	// Verify the client was removed
	if len(broadcastService.clients) != 0 {
		t.Errorf("Expected 0 clients, got %d", len(broadcastService.clients))
	}

	// Verify the channel was closed
	select {
	case _, ok := <-clientChan:
		if ok {
			t.Error("Expected channel to be closed")
		}
	default:
		t.Error("Expected channel to be closed but it's still open")
	}
}

func TestBroadcastService_BroadcastUpdates(t *testing.T) {
	memStore := store.NewMemoryStore(10)
	updateChan := make(chan domain.PriceUpdateEvent, 10)
	broadcastService := NewBroadcastService(memStore, updateChan)

	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	broadcastService.Start(ctx)

	// Subscribe some clients
	client1 := broadcastService.SubscribeClient()
	client2 := broadcastService.SubscribeClient()

	// Send an update
	testUpdate := domain.PriceUpdateEvent{
		Timestamp: time.Now().Unix(),
		Price:     60000.0,
	}
	updateChan <- testUpdate

	// Wait a bit for the update to be processed
	time.Sleep(50 * time.Millisecond)

	// Check if both clients received the update
	select {
	case update := <-client1:
		if update.Price != 60000.0 {
			t.Errorf("Client 1: Expected price 60000.0, got %.2f", update.Price)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client 1: Timeout waiting for update")
	}

	select {
	case update := <-client2:
		if update.Price != 60000.0 {
			t.Errorf("Client 2: Expected price 60000.0, got %.2f", update.Price)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client 2: Timeout waiting for update")
	}
}

func TestBroadcastService_SSEHandler(t *testing.T) {
	memStore := store.NewMemoryStore(10)
	updateChan := make(chan domain.PriceUpdateEvent, 10)
	broadcastService := NewBroadcastService(memStore, updateChan)

	// Start the service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	broadcastService.Start(ctx)

	// Add some test data to the store
	testEvents := []domain.PriceUpdateEvent{
		{Timestamp: 100, Price: 50000.0},
		{Timestamp: 200, Price: 51000.0},
		{Timestamp: 300, Price: 52000.0},
	}

	for _, event := range testEvents {
		memStore.Store(event)
	}

	// Create a test request with a "since" parameter
	req := httptest.NewRequest("GET", "/prices/stream?since=150", nil)

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Create a context that can be cancelled
	reqCtx, reqCancel := context.WithCancel(req.Context())
	req = req.WithContext(reqCtx)

	// Start the SSE handler in a goroutine
	go func() {
		broadcastService.SSEHandler(w, req)
	}()

	// Send a new update that should be received
	newUpdate := domain.PriceUpdateEvent{
		Timestamp: 400,
		Price:     53000.0,
	}

	// Allow some time for initial history to be sent
	time.Sleep(100 * time.Millisecond)

	// Send the update
	updateChan <- newUpdate

	// Give it some time to process
	time.Sleep(100 * time.Millisecond)

	// Cancel the request context to end the handler
	reqCancel()

	// Check the response
	resp := w.Result()
	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream, got %s", resp.Header.Get("Content-Type"))
	}

	// The response body should contain the events with timestamps >= 150
	body := w.Body.String()

	// Check for historical events (timestamp 200 and 300)
	if !strings.Contains(body, `"timestamp":200`) {
		t.Error("Response missing event with timestamp 200")
	}
	if !strings.Contains(body, `"timestamp":300`) {
		t.Error("Response missing event with timestamp 300")
	}

	// Check for the new event
	if !strings.Contains(body, `"timestamp":400`) {
		t.Error("Response missing new event with timestamp 400")
	}

	// Event with timestamp 100 should not be included
	if strings.Contains(body, `"timestamp":100`) {
		t.Error("Response should not include event with timestamp 100")
	}
}
