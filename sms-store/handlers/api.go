package handlers

import (
    "encoding/json"
    "net/http"
    "sms-store/models" // Import your models
)

// 1. Define the Interface (The Contract)
// Anything that has a GetSMSHistory method can be used as our database!
type SMSStore interface {
    GetSMSHistory(userID string) ([]models.SMSRecord, error)
}

// 2. Create a struct to hold our dependencies
type Server struct {
    DB SMSStore
}

// 3. Attach the handler as a method on the Server struct
func (s *Server) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.PathValue("User_id")

    // ⚡ Notice the change here! We use s.DB instead of the direct database package
    messages, err := s.DB.GetSMSHistory(userID)
    if err != nil {
        http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(messages)
}