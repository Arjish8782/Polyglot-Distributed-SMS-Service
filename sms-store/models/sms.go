package models

import "time"

// SMSRecord represents the data we get from Kafka and save to the database
type SMSRecord struct {
	PhoneNumber string    `json:"phoneNumber" bson:"phoneNumber"`
	Message     string    `json:"message"     bson:"message"`
	Status      string    `json:"status"      bson:"status"`
	RequestId   string    `json:"requestId"   bson:"requestId"`
	Timestamp   string    `json:"timestamp"   bson:"timestamp"`
	CreatedAt   time.Time `json:"createdAt"   bson:"createdAt"`
}