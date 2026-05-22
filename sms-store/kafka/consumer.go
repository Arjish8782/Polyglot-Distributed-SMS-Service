package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sms-store/models"

	"github.com/segmentio/kafka-go"
)

// 1. THE INTERFACE
type SMSWriter interface {
	SaveSMS(record models.SMSRecord) error
}

// 2. THE STRUCT
type Consumer struct {
	DB SMSWriter
}

// 3. THE TESTABLE HELPER METHOD (This is what the test was looking for!)
// We separated the JSON parsing and database saving so we can test it without Kafka
func (c *Consumer) ProcessMessage(msgValue []byte) error {
	var smsRecord models.SMSRecord
	err := json.Unmarshal(msgValue, &smsRecord)
	if err != nil {
		return err // Return an error if JSON is bad
	}

	// Save using our injected interface!
	return c.DB.SaveSMS(smsRecord)
}

// 4. THE MAIN LOOP
func (c *Consumer) Start() {
	// Re-initialize the Kafka Reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka:9092"}, // Make sure this matches your Kafka broker!
		Topic:    "sms_events",                // Make sure this matches your Java producer!
		GroupID:  "sms_consumer_group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	
	defer reader.Close()

	fmt.Println("🎧 Kafka Consumer started, listening for messages...")

	for {
		// Read the raw message from Kafka
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("❌ Error reading message: %v\n", err)
			continue
		}

		// ⚡ Pass the raw JSON to our helper method!
		err = c.ProcessMessage(msg.Value)
		if err != nil {
			log.Printf("❌ Failed to process message: %v\n", err)
			continue
		}

		fmt.Printf("✅ Successfully processed and saved SMS\n")
	}
}