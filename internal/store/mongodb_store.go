package store

import (
	"btc-price-tracker/internal/domain"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBStore implements EventStore using MongoDB with TTL
type MongoDBStore struct {
	client     *mongo.Client
	collection *mongo.Collection
	ttl        time.Duration
}

// MongoDBPriceEvent is the MongoDB document structure
type MongoDBPriceEvent struct {
	Timestamp int64     `bson:"timestamp"`
	Price     float64   `bson:"price"`
	ExpiresAt time.Time `bson:"expiresAt"` // TTL field
}

// NewMongoDBStore creates a new MongoDB-backed event store
func NewMongoDBStore(uri string, dbName string, collectionName string, ttl time.Duration) (*MongoDBStore, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get collection
	collection := client.Database(dbName).Collection(collectionName)

	// Create TTL index on expiresAt field
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, err
	}

	// Create timestamp index for efficient querying
	timestampIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "timestamp", Value: 1}},
	}

	_, err = collection.Indexes().CreateOne(ctx, timestampIndex)
	if err != nil {
		return nil, err
	}

	return &MongoDBStore{
		client:     client,
		collection: collection,
		ttl:        ttl,
	}, nil
}

func (ms *MongoDBStore) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ms.client.Disconnect(ctx)
}

// Store saves a price update event to MongoDB
func (ms *MongoDBStore) Store(event domain.PriceUpdateEvent) {
	// Convert domain event to MongoDB document
	doc := MongoDBPriceEvent{
		Timestamp: event.Timestamp,
		Price:     event.Price,
		ExpiresAt: time.Now().Add(ms.ttl), // TTL field
	}

	// Insert document
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := ms.collection.InsertOne(ctx, doc)
	if err != nil {
		log.Printf("Error storing event: %v", err)
	}
}

// GetEventsSince retrieves events since the given timestamp
func (ms *MongoDBStore) GetEventsSince(timestamp int64) []domain.PriceUpdateEvent {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create filter for events with timestamp >= given timestamp
	filter := bson.M{"timestamp": bson.M{"$gte": timestamp}}

	// Set sort order by timestamp ascending
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := ms.collection.Find(ctx, filter, opts)
	if err != nil {
		return []domain.PriceUpdateEvent{}
	}
	defer cursor.Close(ctx)

	var results []domain.PriceUpdateEvent
	for cursor.Next(ctx) {
		var doc MongoDBPriceEvent
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		results = append(results, domain.PriceUpdateEvent{
			Timestamp: doc.Timestamp,
			Price:     doc.Price,
		})
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Error getting events from MongoDB: %v", err)
	}

	return results
}

// GetLatestEvent retrieves the most recent price update event
func (ms *MongoDBStore) GetLatestEvent() (domain.PriceUpdateEvent, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Sort by timestamp descending and limit to 1 result
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var doc MongoDBPriceEvent
	err := ms.collection.FindOne(ctx, bson.M{}, opts).Decode(&doc)
	if err != nil {
		// No document found or error occurred
		return domain.PriceUpdateEvent{}, false
	}

	return domain.PriceUpdateEvent{
		Timestamp: doc.Timestamp,
		Price:     doc.Price,
	}, true
}
