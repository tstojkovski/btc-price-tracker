package store

import "btc-price-tracker/internal/domain"

type EventStore interface {
	Store(event domain.PriceUpdateEvent)
	GetEventsSince(timestamp int64) []domain.PriceUpdateEvent
	GetLatestEvent() (domain.PriceUpdateEvent, bool)
}
