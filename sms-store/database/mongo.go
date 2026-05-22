package database

import (
	"context"
	"fmt"
	"log"
	"sms-store/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
)

var collection *mongo.Collection

// InitDB connects to MongoDB when the application starts
func InitDB() {
	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// We create a database named "smsdb" and a collection named "messages"
	collection = client.Database("smsdb").Collection("messages")
	fmt.Println("🍃 Connected to MongoDB!")
}

// SaveSMS takes the record from Kafka and inserts it into the database
func SaveSMS(record models.SMSRecord) {
	_, err := collection.InsertOne(context.Background(), record)
	if err != nil {
		log.Printf("Error saving to MongoDB: %v", err)
	} else {
		fmt.Println("💾 Successfully saved SMS to database!")
	}
}

// GetSMSHistory fetches all messages for a specific phone number
func GetSMSHistory(phoneNumber string) ([]models.SMSRecord, error) {
	// 1. Create the filter: match documents where the "phoneNumber" field equals our target
	filter := bson.M{"phoneNumber": phoneNumber}

	// 2. Execute the search
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	// 3. Decode the results into a slice (Go's version of a list/array)
	var messages []models.SMSRecord
	if err = cursor.All(context.Background(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}