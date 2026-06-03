package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time" // ⚡ ADDED: Needed for our pause/sleep timer
	"sms-store/models"

	"github.com/segmentio/kafka-go"
)

var ErrInvalidJSON = errors.New("invalid json format")

type SMSWriter interface {
	SaveSMS(record models.SMSRecord) error
}

type Consumer struct {
	DB SMSWriter
}

func (c *Consumer) ProcessMessage(msgValue []byte) error {
	var smsRecord models.SMSRecord
	err := json.Unmarshal(msgValue, &smsRecord)
	if err != nil {
		return ErrInvalidJSON 
	}
	return c.DB.SaveSMS(smsRecord) 
}

func (c *Consumer) Start() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka:9092"}, 
		Topic:    "sms_events",                
		GroupID:  "sms_consumer_group",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	
	defer reader.Close()

	fmt.Println("🎧 Kafka Consumer started, listening for messages...")

	for {
		msg, err := reader.FetchMessage(context.Background())
		if err != nil {
			log.Printf("❌ Error fetching message: %v\n", err)
			continue
		}

		err = c.ProcessMessage(msg.Value)
		
		// ⚡ THE NEW CIRCUIT BREAKER LOGIC
		if err != nil {
			if err == ErrInvalidJSON {
				log.Printf("⚠️ POISON PILL: Bad data received, skipping. Data: %s\n", string(msg.Value))
				reader.CommitMessages(context.Background(), msg)
				continue 
			}

			// INFRASTRUCTURE FAILURE: Do NOT crash! 
			// Instead, enter a waiting loop until the DB comes back.
			log.Printf("🚨 DATABASE DOWN! Pausing Kafka consumption. Error: %v\n", err)
			
			for {
				log.Printf("⏳ Waiting 5 seconds before retrying...")
				time.Sleep(5 * time.Second) // Pause execution for 5 seconds
				
				// Try to save the exact same message again
				retryErr := c.ProcessMessage(msg.Value)
				
				if retryErr == nil {
					log.Printf("🔌 DATABASE RECONNECTED! Message successfully saved.")
					break // Break out of the infinite retry loop and continue normal operation!
				}
				
				log.Printf("❌ Database still down. Retrying again...")
			}
		}

		// Manual Commit on Success (This runs after the first success, OR after the retry loop finishes)
		err = reader.CommitMessages(context.Background(), msg)
		if err != nil {
			log.Printf("❌ Failed to commit offset to Kafka: %v\n", err)
			continue
		}

		fmt.Printf("✅ Successfully processed, saved to DB, and committed to Kafka\n")
	}
}