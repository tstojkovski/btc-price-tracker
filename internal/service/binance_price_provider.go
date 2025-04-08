package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type BinancePriceProvider struct{}

func NewBinancePriceProvider() *BinancePriceProvider {
	return &BinancePriceProvider{}
}

// FetchPrice retrieves the current Bitcoin price in USD
func (p *BinancePriceProvider) FetchPrice() (float64, error) {
	// Use Binance API
	response, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT")
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	// Parse the JSON response
	var result binancePriceResult
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	// Check if we got data
	if result.Price == "" {
		return 0, errors.New("no price data available")
	}

	// Convert string price to float64
	price, err := strconv.ParseFloat(result.Price, 64)
	if err != nil {
		return 0, err
	}

	return price, nil
}

type binancePriceResult struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}
