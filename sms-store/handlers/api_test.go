package handlers

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "sms-store/models"
)

// 1. Create a Fake Database Struct
type MockDatabase struct {}

// 2. Make it satisfy the SMSStore Interface
func (m *MockDatabase) GetSMSHistory(userID string) ([]models.SMSRecord, error) {
    // We don't connect to Mongo! We just return hardcoded fake data.
    fakeMessages := []models.SMSRecord{
        {PhoneNumber: userID, Message: "Mocked Message", Status: "SUCCESS"},
    }
    return fakeMessages, nil
}

func TestGetMessagesHandler_WithoutDocker(t *testing.T) {
    // 1. ARRANGE: Create our Fake Database and inject it into the Server
    mockDB := &MockDatabase{}
    server := &Server{DB: mockDB} // Dependency Injection!

    req, err := http.NewRequest("GET", "/v1/user/9998887777/messages", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()

    // 2. THE ROUTING TRICK
    mux := http.NewServeMux()
    // Notice we use server.GetMessagesHandler now
    mux.HandleFunc("/v1/user/{User_id}/messages", server.GetMessagesHandler) 

    // 3. ACT
    mux.ServeHTTP(rr, req)

    // 4. ASSERT
    if status := rr.Code; status != http.StatusOK {
        t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    t.Logf("Successfully received Mocked JSON response: %s", rr.Body.String())
}