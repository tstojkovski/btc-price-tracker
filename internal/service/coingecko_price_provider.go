package service

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type CoinGeckoPriceProvider struct{}

func NewCoinGeckoPriceProvider() *CoinGeckoPriceProvider {
	return &CoinGeckoPriceProvider{}
}

func (p *CoinGeckoPriceProvider) FetchPrice() (float64, error) {
	// Use coingecko API
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
	var result coinGeckoPriceResult
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	if result.Status != nil && result.Status.ErrorCode != 0 {
		return 0, errors.New(result.Status.ErrorMessage)
	}

	return result.Bitcoin.USD, nil
}

type coinGeckoPriceResult struct {
	Bitcoin struct {
		USD float64 `json:"usd"`
	} `json:"bitcoin"`
	Status *struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"status"`
}
