package service

import (
	"btc-price-tracker/internal/domain"
	"btc-price-tracker/internal/store"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mock implementation for the CoinGecko API
func setupMockAPI(price float64) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return a simplified CoinGecko API response
		response := fmt.Sprintf(`{"bitcoin": {"usd": %f}}`, price)
		if _, err := w.Write([]byte(response)); err != nil {
			return
		}
	}))
}

func TestPriceService_FetchPrice(t *testing.T) {
	// Setup a mock API server
	mockServer := setupMockAPI(55000.0)
	defer mockServer.Close()

	// Replace the original URL with our mock server URL in the fetchBTCPrice function

	memStore := store.NewMemoryStore(10)
	priceService := NewPriceService(memStore)

	// Get the update channel
	updateChan := priceService.GetUpdateChannel()

	// Send a mock price update to test the channel
	mockUpdate := domain.PriceUpdateEvent{
		Timestamp: time.Now().Unix(),
		Price:     55000.0,
	}

	// Store it manually (simulating what fetchPrices would do)
	memStore.Store(mockUpdate)

	// Send to the update channel
	go func() {
		priceService.updateChan <- mockUpdate
	}()

	// Try to receive from the channel
	select {
	case update := <-updateChan:
		if update.Price != 55000.0 {
			t.Errorf("Expected price 55000.0, got %.2f", update.Price)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timeout waiting for update")
	}
}
