package store

import (
	"log"
	"os"
	"strconv"
	"time"
)

// NewStoreFromConfig creates a store implementation based on configuration
func NewStoreFromConfig() EventStore {
	storeType := os.Getenv("STORE_TYPE")
	log.Printf("Store type", storeType)
	switch storeType {
	case "mongo", "mongodb":
		log.Printf("Using mongodb")
		mongoURI := os.Getenv("MONGO_URI")
		if mongoURI == "" {
			mongoURI = "mongodb://localhost:27017"
		}

		database := os.Getenv("MONGO_DATABASE")
		if database == "" {
			database = "btc_price_tracker"
		}

		collection := os.Getenv("MONGO_COLLECTION")
		if collection == "" {
			collection = "price_updates"
		}

		ttlStr := os.Getenv("MONGO_TTL")
		ttl := 3600 // Default: 1 hour in seconds
		if ttlStr != "" {
			var err error
			ttl, err = strconv.Atoi(ttlStr)
			if err != nil {
				log.Printf("Invalid MONGO_TTL value: %s, using default: %d\n", ttlStr, ttl)
			}
		}

		store, err := NewMongoDBStore(mongoURI, database, collection, time.Duration(ttl))
		if err != nil {
			log.Printf("Failed to create MongoDB store: %v, falling back to memory store\n", err)
			return createMemoryStore()
		}

		log.Printf("Using MongoDB store: %s/%s/%s with TTL: %d seconds\n",
			mongoURI, database, collection, ttl)
		return store

	case "memory", "":
		fallthrough
	default:
		return createMemoryStore()
	}
}

func createMemoryStore() EventStore {
	storeSizeStr := os.Getenv("STORE_SIZE")
	storeSize := 100 // Default value

	if storeSizeStr != "" {
		size, err := strconv.Atoi(storeSizeStr)
		if err != nil {
			log.Printf("Invalid STORE_SIZE value: %s, using default: %d\n", storeSizeStr, storeSize)
		} else {
			storeSize = size
		}
	}

	log.Printf("Using memory store with size: %d\n", storeSize)
	return NewMemoryStore(storeSize)
}
