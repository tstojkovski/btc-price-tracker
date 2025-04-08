package service

import (
	"btc-price-tracker/internal/domain"
	"btc-price-tracker/internal/store"
	"context"
	"log"
	"time"
)

type PriceService struct {
	store         store.EventStore
	updateChan    chan domain.PriceUpdateEvent
	priceProvider PriceProvider
}

func NewPriceService(store store.EventStore, priceProvider PriceProvider) *PriceService {
	return &PriceService{
		store:         store,
		priceProvider: priceProvider,
		updateChan:    make(chan domain.PriceUpdateEvent, 1),
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
			price, err := ps.priceProvider.FetchPrice()
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
