package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sms-store/models"

	"github.com/segmentio/kafka-go"
)

// ⚡ Create a specific error just for bad JSON data
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
		return ErrInvalidJSON // Return our specific Poison Pill error
	}

	return c.DB.SaveSMS(smsRecord) // Return the DB error (if any)
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
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("❌ Error reading message: %v\n", err)
			continue
		}

		err = c.ProcessMessage(msg.Value)
		if err != nil {
			// 1. POISON PILL CHECK: Is it just bad data?
			if err == ErrInvalidJSON {
				log.Printf("⚠️ POISON PILL: Bad data received, skipping. Data: %s\n", string(msg.Value))
				continue // Move on to the next message
			}

			// 2. INFRASTRUCTURE FAILURE: It's a DB error! Crash the consumer!
			log.Fatalf("🚨 CRITICAL: Database failure! Crashing consumer to preserve Kafka offset. Error: %v\n", err)
		}

		fmt.Printf("✅ Successfully processed and saved SMS\n")
	}
}