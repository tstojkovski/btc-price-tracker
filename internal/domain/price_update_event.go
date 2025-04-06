package domain

type PriceUpdateEvent struct {
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
}
