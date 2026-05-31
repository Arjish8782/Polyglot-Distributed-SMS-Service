package kafka

import (
	"fmt"
	"testing"
	"sms-store/models"
)

type MockConsumerDB struct {
	SavedRecord models.SMSRecord 
}

func (m *MockConsumerDB) SaveSMS(record models.SMSRecord) error {
	m.SavedRecord = record 
	return nil
}

// ⚡ NEW: A mock database that always fails!
type MockFailingDB struct {}

func (m *MockFailingDB) SaveSMS(record models.SMSRecord) error {
	return fmt.Errorf("connection timeout")
}

func TestProcessMessage_Success(t *testing.T) {
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	fakeKafkaMessage := []byte(`{"phoneNumber":"1112223333","message":"Test from Kafka!","status":"SUCCESS"}`)
	err := consumer.ProcessMessage(fakeKafkaMessage)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if mockDB.SavedRecord.PhoneNumber != "1112223333" {
		t.Errorf("Expected phone number 1112223333, got %s", mockDB.SavedRecord.PhoneNumber)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	badJSON := []byte(`{this_is_not_json}`)
	err := consumer.ProcessMessage(badJSON)

	// ⚡ Check specifically for our custom Poison Pill error
	if err != ErrInvalidJSON {
		t.Errorf("Expected ErrInvalidJSON for bad data, but got: %v", err)
	}
}

// ⚡ NEW TEST: Ensure DB errors are bubbled up properly so Start() can crash
func TestProcessMessage_DatabaseFailure(t *testing.T) {
	mockDB := &MockFailingDB{} // Use the failing DB!
	consumer := &Consumer{DB: mockDB}

	goodJSON := []byte(`{"phoneNumber":"1112223333","message":"Test!","status":"SUCCESS"}`)
	err := consumer.ProcessMessage(goodJSON)

	// We expect a database error, not a JSON error
	if err == nil {
		t.Fatalf("Expected a database error, but got nil")
	}
	if err == ErrInvalidJSON {
		t.Fatalf("Expected a database error, but got ErrInvalidJSON")
	}
}