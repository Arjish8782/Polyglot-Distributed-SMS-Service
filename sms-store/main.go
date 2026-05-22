package main

import (
	"fmt"
	"net/http"
	"sms-store/database"
	"sms-store/handlers"
	"sms-store/kafka"
	"sms-store/models"
)

// 1. Our single "Real" database struct
type RealMongoDatabase struct{}

// 2a. Satisfy the API's Interface (from earlier)
func (r *RealMongoDatabase) GetSMSHistory(userID string) ([]models.SMSRecord, error) {
	return database.GetSMSHistory(userID)
}

// 2b. Satisfy the Consumer's Interface (NEW!)
func (r *RealMongoDatabase) SaveSMS(record models.SMSRecord) error {
	database.SaveSMS(record) // 1. Call your original function
	return nil               // 2. Return a blank "no error" to satisfy the interface
}

func main() {
	// Connect to the actual MongoDB via Docker
	database.InitDB()

	// 3. Create ONE instance of our real database
	realDB := &RealMongoDatabase{}

	// 4. DEPENDENCY INJECTION: Build the Kafka Consumer and start it
	messageConsumer := &kafka.Consumer{DB: realDB}
	go messageConsumer.Start() // Notice we call the method now!

	// 5. DEPENDENCY INJECTION: Build the HTTP Server
	server := &handlers.Server{DB: realDB}
	http.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler)

	// Start the web server
	fmt.Println("🚀 SMS Store Web Server running on http://localhost:8081...")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}