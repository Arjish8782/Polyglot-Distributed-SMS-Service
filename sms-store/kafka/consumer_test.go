package kafka

import (
	"testing"
	"sms-store/models"
)

// 1. Create a Fake Database just for the Consumer Test
type MockConsumerDB struct {
	SavedRecord models.SMSRecord // We use this to spy on what got saved!
}

// 2. Satisfy the SMSWriter Interface
func (m *MockConsumerDB) SaveSMS(record models.SMSRecord) error {
	m.SavedRecord = record // Capture the record so the test can inspect it
	return nil
}

func TestProcessMessage_Success(t *testing.T) {
	// ARRANGE: Build the consumer with the fake database
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	// Create a fake JSON byte array that looks exactly like what Java sends
	fakeKafkaMessage := []byte(`{"phoneNumber":"1112223333","message":"Test from Kafka!","status":"SUCCESS"}`)

	// ACT: Pass the fake JSON to our processing method
	err := consumer.ProcessMessage(fakeKafkaMessage)

	// ASSERT: Check that no errors occurred
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// ASSERT: Check that our fake database captured the correct data
	if mockDB.SavedRecord.PhoneNumber != "1112223333" {
		t.Errorf("Expected phone number 1112223333, got %s", mockDB.SavedRecord.PhoneNumber)
	}
	if mockDB.SavedRecord.Message != "Test from Kafka!" {
		t.Errorf("Expected message 'Test from Kafka!', got %s", mockDB.SavedRecord.Message)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	// Send broken, malformed JSON
	badJSON := []byte(`{this_is_not_json}`)

	// ACT
	err := consumer.ProcessMessage(badJSON)

	// ASSERT: We EXPECT an error here because the JSON is bad
	if err == nil {
		t.Error("Expected an error for bad JSON, but got nil")
	}
}