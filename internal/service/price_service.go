package service

import (
	"btc-price-tracker/internal/domain"
	"btc-price-tracker/internal/store"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type PriceService struct {
	store      store.EventStore
	updateChan chan domain.PriceUpdateEvent
}

func NewPriceService(store store.EventStore) *PriceService {
	return &PriceService{
		store:      store,
		updateChan: make(chan domain.PriceUpdateEvent, 1),
	}
}

func (ps *PriceService) Start(ctx context.Context) {
	go ps.fetchPrices(ctx)
}

func (ps *PriceService) GetUpdateChannel() <-chan domain.PriceUpdateEvent {
	return ps.updateChan
}

// fetchPrices periodically fetches BTC/USD prices
func (ps *PriceService) fetchPrices(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			price, err := fetchBTCPrice()
			if err != nil {
				log.Printf("Error fetching price: %v", err)
				continue
			}

			update := domain.PriceUpdateEvent{
				Timestamp: time.Now().Unix(),
				Price:     price,
			}

			ps.store.Store(update)

			select {
			case ps.updateChan <- update:
				// Successfully sent update
			default:
				// Channel buffer is full, log and move on
				log.Println("Update channel buffer full, notification skipped")
			}

			log.Printf("New price: $%.2f at %v", update.Price, time.Unix(update.Timestamp, 0))

		case <-ctx.Done():
			log.Println("Stopping price fetcher")
			return
		}
	}
}

func fetchBTCPrice() (float64, error) {
	// Use coingecko for now
	response, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd")
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	// Parse the JSON response
	var result CoinGeckoBTCPriceResult
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if result.Status != nil && result.Status.ErrorCode != 0 {
		return 0, errors.New(result.Status.ErrorMessage)
	}

	return result.Bitcoin.USD, nil
}

type CoinGeckoBTCPriceResult struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
	} `json:"bitcoin"`
	Status *struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}
