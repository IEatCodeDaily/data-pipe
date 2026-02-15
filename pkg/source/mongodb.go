package source

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/IEatCodeDaily/data-pipe/pkg/pipeline"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBSource implements the Source interface for MongoDB
type MongoDBSource struct {
	uri        string
	database   string
	collection string
	client     *mongo.Client
	logger     *log.Logger
}

// NewMongoDBSource creates a new MongoDB source
func NewMongoDBSource(uri, database, collection string, logger *log.Logger) *MongoDBSource {
	if logger == nil {
		logger = log.Default()
	}
	return &MongoDBSource{
		uri:        uri,
		database:   database,
		collection: collection,
		logger:     logger,
	}
}

// Connect establishes connection to MongoDB
func (m *MongoDBSource) Connect(ctx context.Context) error {
	m.logger.Printf("Connecting to MongoDB: %s", m.uri)

	clientOptions := options.Client().ApplyURI(m.uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	m.client = client
	m.logger.Println("Successfully connected to MongoDB")
	return nil
}

// Read reads change events from MongoDB using change streams
func (m *MongoDBSource) Read(ctx context.Context) (<-chan pipeline.Event, <-chan error) {
	events := make(chan pipeline.Event)
	errors := make(chan error)

	go func() {
		defer close(events)
		defer close(errors)

		collection := m.client.Database(m.database).Collection(m.collection)

		// Create a change stream
		pipeline := mongo.Pipeline{}
		opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

		m.logger.Printf("Starting change stream for %s.%s", m.database, m.collection)
		stream, err := collection.Watch(ctx, pipeline, opts)
		if err != nil {
			errors <- fmt.Errorf("failed to create change stream: %w", err)
			return
		}
		defer stream.Close(ctx)

		for stream.Next(ctx) {
			var changeDoc bson.M
			if err := stream.Decode(&changeDoc); err != nil {
				errors <- fmt.Errorf("failed to decode change event: %w", err)
				continue
			}

			event := m.convertChangeEvent(changeDoc)
			events <- event
		}

		if err := stream.Err(); err != nil {
			errors <- fmt.Errorf("change stream error: %w", err)
		}
	}()

	return events, errors
}

// convertChangeEvent converts MongoDB change stream event to pipeline event
func (m *MongoDBSource) convertChangeEvent(changeDoc bson.M) pipeline.Event {
	event := pipeline.Event{
		Source:     "mongodb",
		Database:   m.database,
		Collection: m.collection,
		Timestamp:  time.Now(),
	}

	if id, ok := changeDoc["_id"]; ok {
		event.ID = fmt.Sprintf("%v", id)
	}

	if opType, ok := changeDoc["operationType"].(string); ok {
		event.Operation = opType
	}

	if fullDoc, ok := changeDoc["fullDocument"].(bson.M); ok {
		event.Data = convertBSONToMap(fullDoc)
	}

	if updateDesc, ok := changeDoc["updateDescription"].(bson.M); ok {
		if updatedFields, ok := updateDesc["updatedFields"].(bson.M); ok {
			if event.Data == nil {
				event.Data = make(map[string]interface{})
			}
			for k, v := range convertBSONToMap(updatedFields) {
				event.Data[k] = v
			}
		}
	}

	return event
}

// convertBSONToMap converts BSON document to map
func convertBSONToMap(doc bson.M) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range doc {
		result[k] = v
	}
	return result
}

// Close closes the MongoDB connection
func (m *MongoDBSource) Close() error {
	if m.client != nil {
		m.logger.Println("Closing MongoDB connection")
		return m.client.Disconnect(context.Background())
	}
	return nil
}
