package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sms-store/models"
	"strconv"
)

type SMSStore interface {
	GetSMSHistory(userID string, page, limit int) ([]models.SMSRecord, error)
}

type Server struct {
	DB SMSStore
}

func (s *Server) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("User_id")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	messages, err := s.DB.GetSMSHistory(userID, page, limit)
	if err != nil {
		log.Printf("ERROR: GetSMSHistory failed for user %s: %v", userID, err)
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
