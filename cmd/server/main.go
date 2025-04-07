package main

import (
	"btc-price-tracker/internal/service"
	"btc-price-tracker/internal/store"
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	storeSizeStr := os.Getenv("STORE_SIZE")
	storeSize := 100 // Default value

	// Convert string to integer if the environment variable exists
	if storeSizeStr != "" {
		size, err := strconv.Atoi(storeSizeStr)
		if err != nil {
			log.Printf("Warning: Invalid STORE_SIZE value '%s', using default: %d\n", storeSizeStr, storeSize)
		} else {
			storeSize = size
		}
	}

	log.Printf("Using store size: %d\n", storeSize)

	store := store.NewMemoryStore(storeSize)

	priceService := service.NewPriceService(store)
	priceService.Start(ctx)

	broadcastService := service.NewBroadcastService(store, priceService.GetUpdateChannel())
	broadcastService.Start(ctx)

	mux := http.NewServeMux()

	mux.HandleFunc("/prices/stream", broadcastService.SSEHandler)

	staticDir := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticDir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		// Read the index.html file
		indexPath := filepath.Join("static", "index.html")
		indexContent, err := os.ReadFile(indexPath)
		if err != nil {
			log.Printf("Error reading index.html: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if _, err = w.Write(indexContent); err != nil {
			log.Printf("Error writing: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	server := &http.Server{
		Addr:              ":8082",
		Handler:           mux,
		ReadHeaderTimeout: time.Second * 30,
	}

	log.Println("Server starting on :8082")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}
