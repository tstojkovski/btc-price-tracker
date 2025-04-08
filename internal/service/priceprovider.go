package service

type PriceProvider interface {
	FetchPrice() (float64, error)
}
