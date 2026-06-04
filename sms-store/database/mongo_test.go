package database

import (
	"testing"
	"sms-store/models"
)

func TestSaveAndGetSMS(t *testing.T) {
	InitDB()

	dummyPhone := "5554443333"
	testRecord := models.SMSRecord{
		PhoneNumber: dummyPhone,
		Message:     "Integration Test Message",
		Status:      "SUCCESS",
	}

	if err := SaveSMS(testRecord); err != nil {
		t.Fatalf("SaveSMS returned unexpected error: %v", err)
	}

	history, err := GetSMSHistory(dummyPhone, 1, 20)
	if err != nil {
		t.Fatalf("Failed to retrieve SMS history: %v", err)
	}

	found := false
	for _, msg := range history {
		if msg.Message == "Integration Test Message" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Saved message was not found in the retrieved history!")
	}
}