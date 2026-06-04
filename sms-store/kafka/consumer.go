package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"sms-store/models"
	"time"

	"github.com/segmentio/kafka-go"
)

var ErrInvalidJSON = errors.New("invalid json format")

// MessageReader abstracts kafka.Reader so Start() can be tested without a real broker.
type MessageReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type SMSWriter interface {
	SaveSMS(record models.SMSRecord) error
}

type Consumer struct {
	DB     SMSWriter
	Reader MessageReader // injected in tests; nil → Start() creates the real kafka.Reader
}

// backoffDuration returns an exponentially increasing delay with ±25% jitter, capped at 60s.
func backoffDuration(attempt int) time.Duration {
	base := time.Second * (1 << attempt) // 1s, 2s, 4s, 8s, ...
	if base > 60*time.Second {
		base = 60 * time.Second
	}
	jitter := time.Duration(rand.Int63n(int64(base) / 2))
	return base + jitter - base/4
}

func (c *Consumer) ProcessMessage(msgValue []byte) error {
	var smsRecord models.SMSRecord
	if err := json.Unmarshal(msgValue, &smsRecord); err != nil {
		return ErrInvalidJSON
	}
	// FAILED vendor events stay in Kafka as audit trail but are not persisted to MongoDB
	if smsRecord.Status == "FAILED" {
		log.Printf("Skipping FAILED event for %s", smsRecord.PhoneNumber)
		return nil
	}
	return c.DB.SaveSMS(smsRecord)
}

func (c *Consumer) Start(ctx context.Context) {
	reader := c.Reader
	if reader == nil {
		brokers := os.Getenv("KAFKA_BROKERS")
		if brokers == "" {
			brokers = "kafka:9092"
		}
		reader = kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{brokers},
			Topic:    "sms_events",
			GroupID:  "sms_consumer_group",
			MinBytes: 10e3,
			MaxBytes: 10e6,
		})
	}
	defer reader.Close()

	log.Println("Kafka Consumer started, listening for messages...")

	fetchErrCount := 0
	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			// Context cancelled = intentional shutdown, exit cleanly
			if ctx.Err() != nil {
				log.Println("Consumer shutting down")
				return
			}
			fetchErrCount++
			sleep := backoffDuration(fetchErrCount - 1)
			log.Printf("Error fetching message (attempt %d): %v. Retrying in %v", fetchErrCount, err, sleep)
			select {
			case <-ctx.Done():
				return
			case <-time.After(sleep):
			}
			continue
		}
		fetchErrCount = 0

		err = c.ProcessMessage(msg.Value)
		if err != nil {
			if errors.Is(err, ErrInvalidJSON) {
				log.Printf("POISON PILL: Bad data, skipping. Data: %s", string(msg.Value))
				reader.CommitMessages(ctx, msg)
				continue
			}

			// Infrastructure failure — retry with backoff until DB recovers
			log.Printf("DATABASE ERROR: %v. Pausing consumption.", err)
			for attempt := 0; ; attempt++ {
				sleep := backoffDuration(attempt)
				log.Printf("Retrying in %v (attempt %d)...", sleep, attempt+1)
				select {
				case <-ctx.Done():
					log.Println("Consumer shutting down during DB retry")
					return
				case <-time.After(sleep):
				}
				if retryErr := c.ProcessMessage(msg.Value); retryErr == nil {
					log.Printf("DATABASE RECONNECTED after %d retries.", attempt+1)
					break
				}
			}
		}

		if err = reader.CommitMessages(ctx, msg); err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("Failed to commit offset: %v", err)
			continue
		}
		log.Printf("Message processed and committed to Kafka")
	}
}
