package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sms-store/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		uri = "mongodb://mongodb:27017"
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	// Ping verifies actual network connectivity, not just client creation
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to connect to MongoDB (ping failed): %v", err)
	}

	collection = client.Database("smsdb").Collection("messages")

	// Create compound index on (phoneNumber ASC, createdAt DESC) for fast paginated queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "phoneNumber", Value: 1},
			{Key: "createdAt", Value: -1},
		},
		Options: options.Index().SetName("phoneNumber_createdAt_idx"),
	}
	if _, err = collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		log.Fatalf("Failed to create MongoDB index: %v", err)
	}

	log.Println("Connected to MongoDB")
	fmt.Println("🍃 Connected to MongoDB!")
}

func SaveSMS(record models.SMSRecord) error {
	if collection == nil {
		return fmt.Errorf("database not initialized")
	}
	record.CreatedAt = time.Now().UTC()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, record)
	return err
}

func GetSMSHistory(phoneNumber string, page, limit int) ([]models.SMSRecord, error) {
	if collection == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := int64((page - 1) * limit)
	opts := options.Find().
		SetSkip(skip).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := collection.Find(ctx, bson.M{"phoneNumber": phoneNumber}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var messages []models.SMSRecord
	if err = cursor.All(ctx, &messages); err != nil {
		return nil, err
	}
	return messages, nil
}