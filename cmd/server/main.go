package main

import (
	"btc-price-tracker/internal/service"
	"btc-price-tracker/internal/store"
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	serverPort      = ":8082"
	staticDirPath   = "./static"
	indexHtmlPath   = "static/index.html"
	providerEnvVar  = "PRICE_PROVIDER"
	providerBinance = "BINANCE"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	store := store.NewStoreFromConfig()
	priceProvider := initializePriceProvider()
	priceService := service.NewPriceService(store, priceProvider)
	broadcastService := service.NewBroadcastService(store, priceService.GetUpdateChannel())

	// Start services
	priceService.Start(ctx)
	broadcastService.Start(ctx)

	// Setup and start HTTP server
	server := setupServer(broadcastService)

	log.Printf("Server starting on %s", serverPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// initializePriceProvider creates a price provider based on environment configuration
func initializePriceProvider() service.PriceProvider {
	priceProviderStr := os.Getenv(providerEnvVar)
	if priceProviderStr == providerBinance {
		return service.NewBinancePriceProvider()
	}
	return service.NewCoinGeckoPriceProvider()
}

// setupServer configures the HTTP server and routes
func setupServer(broadcastService *service.BroadcastService) *http.Server {
	mux := http.NewServeMux()

	// Setup routes
	mux.HandleFunc("/prices/stream", broadcastService.SSEHandler)
	setupStaticRoutes(mux)

	return &http.Server{
		Addr:              serverPort,
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 30,
	}
}

// setupStaticRoutes configures routes for static content
func setupStaticRoutes(mux *http.ServeMux) {
	// Static file server
	staticDir := http.FileServer(http.Dir(staticDirPath))
	mux.Handle("/static/", http.StripPrefix("/static/", staticDir))

	// Root handler for serving index.html
	mux.HandleFunc("/", serveIndexHandler)
}

// serveIndexHandler serves the index.html file
func serveIndexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	indexContent, err := os.ReadFile(indexHtmlPath)
	if err != nil {
		log.Printf("Error reading index.html: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err = w.Write(indexContent); err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
