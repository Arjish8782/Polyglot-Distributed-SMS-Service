package database

import (
	"testing"
	"time"
	"sms-store/models" 
)

func TestSaveAndGetSMS(t *testing.T) {
	// 1. Connect to the local MongoDB container
	InitDB()

	// 2. Create a dummy SMS record
	dummyPhone := "5554443333"
	testRecord := models.SMSRecord{
		PhoneNumber: dummyPhone,
		Message:     "Integration Test Message",
		Status:      "SUCCESS",
	}

	// 3. ACT: Attempt to save it to the database
	// REMOVED the "err :=" here because SaveSMS doesn't return one!
	SaveSMS(testRecord)

	// 4. ACT: Attempt to retrieve it
	time.Sleep(1 * time.Second) // Wait for Mongo to write
	
	// GetSMSHistory DOES return an error, so we keep this one as is!
	history, err := GetSMSHistory(dummyPhone)
	if err != nil {
		t.Fatalf("Failed to retrieve SMS history: %v", err)
	}

	// 5. ASSERT: Verify the data matches
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