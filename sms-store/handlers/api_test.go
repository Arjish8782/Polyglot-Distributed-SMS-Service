package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sms-store/models"
	"testing"
)

type MockDatabase struct{}

func (m *MockDatabase) GetSMSHistory(userID string, page, limit int) ([]models.SMSRecord, error) {
	fakeMessages := []models.SMSRecord{
		{PhoneNumber: userID, Message: "Mocked Message", Status: "SUCCESS"},
	}
	return fakeMessages, nil
}

type MockFailingDatabase struct{}

func (m *MockFailingDatabase) GetSMSHistory(userID string, page, limit int) ([]models.SMSRecord, error) {
	return nil, fmt.Errorf("simulated DB failure")
}

func TestGetMessagesHandler_Success(t *testing.T) {
	mockDB := &MockDatabase{}
	server := &Server{DB: mockDB}

	req, err := http.NewRequest("GET", "/v1/user/9998887777/messages", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler)
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200, got %v", status)
	}
}

func TestGetMessagesHandler_DBError(t *testing.T) {
	mockDB := &MockFailingDatabase{}
	server := &Server{DB: mockDB}

	req, err := http.NewRequest("GET", "/v1/user/9998887777/messages", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler)
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %v", status)
	}
}

func TestGetMessagesHandler_PaginationParams(t *testing.T) {
	mockDB := &MockDatabase{}
	server := &Server{DB: mockDB}

	req, err := http.NewRequest("GET", "/v1/user/9998887777/messages?page=2&limit=5", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler)
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected 200 with pagination params, got %v", status)
	}
}
