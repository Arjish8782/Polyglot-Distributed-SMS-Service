package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sms-store/database"
	"sms-store/handlers"
	"sms-store/kafka"
	"sms-store/models"
	"syscall"
	"time"
)

type RealMongoDatabase struct{}

func (r *RealMongoDatabase) GetSMSHistory(userID string, page, limit int) ([]models.SMSRecord, error) {
	return database.GetSMSHistory(userID, page, limit)
}

func (r *RealMongoDatabase) SaveSMS(record models.SMSRecord) error {
	return database.SaveSMS(record)
}

func main() {
	database.InitDB()

	// Root context: cancelled on SIGTERM or Ctrl-C
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, os.Interrupt)
	defer cancel()

	realDB := &RealMongoDatabase{}
	messageConsumer := &kafka.Consumer{DB: realDB}

	// Supervised consumer goroutine: recovers from panics and restarts until shutdown
	go func() {
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("PANIC in consumer goroutine: %v. Restarting...", r)
					}
				}()
				messageConsumer.Start(ctx)
			}()

			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()

	mux := http.NewServeMux()
	server := &handlers.Server{DB: realDB}
	mux.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler)

	httpServer := &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	go func() {
		log.Println("SMS Store running on :8081")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
	log.Println("Shutdown complete")
}
