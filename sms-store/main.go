package main

import (
	"fmt"
	"net/http"
	"sms-store/database"
	"sms-store/handlers"
	"sms-store/kafka"
	"sms-store/models"
)

type RealMongoDatabase struct{}

func (r *RealMongoDatabase) GetSMSHistory(userID string) ([]models.SMSRecord, error) {
	return database.GetSMSHistory(userID)
}

// ⚡ THE FIX: Actually return the error from MongoDB instead of hardcoding nil!
// (Note: Make sure your database.SaveSMS function is set up to return an error)
func (r *RealMongoDatabase) SaveSMS(record models.SMSRecord) error {
	return database.SaveSMS(record) 
}

func main() {
	database.InitDB()

	realDB := &RealMongoDatabase{}

	messageConsumer := &kafka.Consumer{DB: realDB}
	go messageConsumer.Start()

	server := &handlers.Server{DB: realDB}
	http.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler)

	fmt.Println("🚀 SMS Store Web Server running on http://localhost:8081...")
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}