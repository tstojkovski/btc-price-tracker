package service

import (
	"btc-price-tracker/internal/domain"
	"btc-price-tracker/internal/store"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type BroadcastService struct {
	store      store.EventStore
	clients    map[chan domain.PriceUpdateEvent]bool
	mutex      sync.RWMutex
	updateChan <-chan domain.PriceUpdateEvent
}

func NewBroadcastService(store store.EventStore, updateChan <-chan domain.PriceUpdateEvent) *BroadcastService {
	return &BroadcastService{
		store:      store,
		clients:    make(map[chan domain.PriceUpdateEvent]bool),
		updateChan: updateChan,
	}
}

func (bs *BroadcastService) Start(ctx context.Context) {
	go bs.broadcastUpdates(ctx)
}

func (bs *BroadcastService) broadcastUpdates(ctx context.Context) {
	for {
		select {
		case update := <-bs.updateChan:
			bs.broadcastToAllClients(update)
		case <-ctx.Done():
			log.Println("Stopping broadcast service")
			return
		}
	}
}

func (bs *BroadcastService) broadcastToAllClients(update domain.PriceUpdateEvent) {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	for client := range bs.clients {
		select {
		case client <- update:
			// Successfully sent update
		default:
			// Client channel buffer full, skip this client
		}
	}
}

func (bs *BroadcastService) SubscribeClient() chan domain.PriceUpdateEvent {
	clientChan := make(chan domain.PriceUpdateEvent, 10) // Buffer to handle some backpressure

	bs.mutex.Lock()
	bs.clients[clientChan] = true
	bs.mutex.Unlock()

	return clientChan
}

func (bs *BroadcastService) UnsubscribeClient(clientChan chan domain.PriceUpdateEvent) {
	bs.mutex.Lock()
	delete(bs.clients, clientChan)
	close(clientChan)
	bs.mutex.Unlock()
}

func (bs *BroadcastService) SSEHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	sinceStr := r.URL.Query().Get("since")
	var lastTimestamp int64 = 0

	if sinceStr != "" {
		since, err := strconv.ParseInt(sinceStr, 10, 64)
		if err == nil {
			lastTimestamp = since

			// Send historical updates
			events := bs.store.GetEventsSince(since)

			log.Println("Loaded historical events: ", len(events))
			for _, event := range events {
				data, err := json.Marshal(event)
				if err != nil {
					log.Printf("Error marshaling event: %v", err)
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
				lastTimestamp = event.Timestamp
			}
		}
	} else {
		// If no since parameter, send the latest event if available
		if latestEvent, exists := bs.store.GetLatestEvent(); exists {
			data, err := json.Marshal(latestEvent)
			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
				lastTimestamp = latestEvent.Timestamp
			}
		}
	}

	clientChan := bs.SubscribeClient()

	ctx := r.Context()

	go func() {
		<-ctx.Done()
		bs.UnsubscribeClient(clientChan)
	}()

	for event := range clientChan {
		if event.Timestamp > lastTimestamp {
			data, err := json.Marshal(event)
			if err != nil {
				log.Printf("Error marshaling event: %v", err)
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			lastTimestamp = event.Timestamp
		}
	}
}
