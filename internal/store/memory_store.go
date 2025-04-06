package store

import (
	"btc-price-tracker/internal/domain"
	"sync"
)

type MemoryStore struct {
	events    []domain.PriceUpdateEvent
	mu        sync.RWMutex
	capacity  int
	nextIndex int
	size      int
}

func NewMemoryStore(capacity int) *MemoryStore {
	return &MemoryStore{
		events:    make([]domain.PriceUpdateEvent, capacity),
		capacity:  capacity,
		nextIndex: 0,
		size:      0,
	}
}

func (ms *MemoryStore) Store(event domain.PriceUpdateEvent) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.events[ms.nextIndex] = event

	ms.nextIndex = (ms.nextIndex + 1) % ms.capacity
	if ms.size < ms.capacity {
		ms.size++
	}
}

func (ms *MemoryStore) GetEventsSince(timestamp int64) []domain.PriceUpdateEvent {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	result := make([]domain.PriceUpdateEvent, 0, ms.size)

	if ms.size == 0 {
		return result
	}

	startIdx := ms.nextIndex - ms.size
	if startIdx < 0 {
		startIdx += ms.capacity
	}

	for i := 0; i < ms.size; i++ {
		idx := (startIdx + i) % ms.capacity
		if ms.events[idx].Timestamp >= timestamp {
			result = append(result, ms.events[idx])
		}
	}

	return result
}

func (ms *MemoryStore) GetLatestEvent() (domain.PriceUpdateEvent, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if ms.size == 0 {
		return domain.PriceUpdateEvent{}, false
	}

	latestIdx := ms.nextIndex - 1
	if latestIdx < 0 {
		latestIdx += ms.capacity
	}

	return ms.events[latestIdx], true
}
