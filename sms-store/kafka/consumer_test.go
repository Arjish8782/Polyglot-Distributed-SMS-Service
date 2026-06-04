package kafka

import (
	"context"
	"fmt"
	"sms-store/models"
	"testing"

	"github.com/segmentio/kafka-go"
)

// ── Mock DB ──────────────────────────────────────────────────────────────────

type MockConsumerDB struct {
	SavedRecord models.SMSRecord
	WasCalled   bool
	CallCount   int
}

func (m *MockConsumerDB) SaveSMS(record models.SMSRecord) error {
	m.WasCalled = true
	m.CallCount++
	m.SavedRecord = record
	return nil
}

type MockFailingDB struct{}

func (m *MockFailingDB) SaveSMS(record models.SMSRecord) error {
	return fmt.Errorf("connection timeout")
}

// MockFailThenSucceedDB fails the first N calls then succeeds.
type MockFailThenSucceedDB struct {
	FailCount int
	calls     int
}

func (m *MockFailThenSucceedDB) SaveSMS(record models.SMSRecord) error {
	m.calls++
	if m.calls <= m.FailCount {
		return fmt.Errorf("db unavailable (call %d)", m.calls)
	}
	return nil
}

// ── Mock Reader ───────────────────────────────────────────────────────────────

// MockReader plays back a fixed list of messages then blocks until ctx is cancelled.
type MockReader struct {
	messages  []kafka.Message
	fetchIdx  int
	Committed []kafka.Message
}

func (m *MockReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if m.fetchIdx < len(m.messages) {
		msg := m.messages[m.fetchIdx]
		m.fetchIdx++
		return msg, nil
	}
	// All messages consumed — block until context is cancelled
	<-ctx.Done()
	return kafka.Message{}, ctx.Err()
}

func (m *MockReader) CommitMessages(_ context.Context, msgs ...kafka.Message) error {
	m.Committed = append(m.Committed, msgs...)
	return nil
}

func (m *MockReader) Close() error { return nil }

// ── ProcessMessage unit tests ────────────────────────────────────────────────

func TestProcessMessage_Success(t *testing.T) {
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	msg := []byte(`{"phoneNumber":"1112223333","message":"Test from Kafka!","status":"SUCCESS"}`)
	if err := consumer.ProcessMessage(msg); err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if mockDB.SavedRecord.PhoneNumber != "1112223333" {
		t.Errorf("Expected phone 1112223333, got %s", mockDB.SavedRecord.PhoneNumber)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	if err := consumer.ProcessMessage([]byte(`{this_is_not_json}`)); err != ErrInvalidJSON {
		t.Errorf("Expected ErrInvalidJSON, got: %v", err)
	}
}

func TestProcessMessage_DatabaseFailure(t *testing.T) {
	consumer := &Consumer{DB: &MockFailingDB{}}

	err := consumer.ProcessMessage([]byte(`{"phoneNumber":"1112223333","message":"Test!","status":"SUCCESS"}`))
	if err == nil {
		t.Fatalf("Expected a database error, got nil")
	}
	if err == ErrInvalidJSON {
		t.Fatalf("Expected a database error, got ErrInvalidJSON")
	}
}

func TestProcessMessage_FailedStatus(t *testing.T) {
	mockDB := &MockConsumerDB{}
	consumer := &Consumer{DB: mockDB}

	err := consumer.ProcessMessage([]byte(`{"phoneNumber":"1112223333","message":"Test!","status":"FAILED"}`))
	if err != nil {
		t.Fatalf("Expected nil for FAILED event, got: %v", err)
	}
	if mockDB.WasCalled {
		t.Errorf("SaveSMS should NOT be called for FAILED events")
	}
}

// ── Start() loop integration tests ───────────────────────────────────────────

func TestStart_SuccessfulMessage(t *testing.T) {
	mockDB := &MockConsumerDB{}
	reader := &MockReader{
		messages: []kafka.Message{
			{Value: []byte(`{"phoneNumber":"111","message":"hello","status":"SUCCESS"}`)},
		},
	}
	consumer := &Consumer{DB: mockDB, Reader: reader}

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after the one message is consumed and committed
	go func() {
		for len(reader.Committed) == 0 {
		}
		cancel()
	}()
	consumer.Start(ctx)

	if mockDB.CallCount != 1 {
		t.Errorf("Expected SaveSMS called once, got %d", mockDB.CallCount)
	}
	if len(reader.Committed) != 1 {
		t.Errorf("Expected 1 committed message, got %d", len(reader.Committed))
	}
}

func TestStart_PoisonPillSkipped(t *testing.T) {
	mockDB := &MockConsumerDB{}
	reader := &MockReader{
		messages: []kafka.Message{
			{Value: []byte(`{not valid json}`)},
		},
	}
	consumer := &Consumer{DB: mockDB, Reader: reader}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for len(reader.Committed) == 0 {
		}
		cancel()
	}()
	consumer.Start(ctx)

	if mockDB.WasCalled {
		t.Errorf("SaveSMS should NOT be called for a poison pill")
	}
	if len(reader.Committed) != 1 {
		t.Errorf("Poison pill offset should be committed (got %d commits)", len(reader.Committed))
	}
}

func TestStart_FailedEventSkipped(t *testing.T) {
	mockDB := &MockConsumerDB{}
	reader := &MockReader{
		messages: []kafka.Message{
			{Value: []byte(`{"phoneNumber":"111","message":"hi","status":"FAILED"}`)},
		},
	}
	consumer := &Consumer{DB: mockDB, Reader: reader}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for len(reader.Committed) == 0 {
		}
		cancel()
	}()
	consumer.Start(ctx)

	if mockDB.WasCalled {
		t.Errorf("SaveSMS should NOT be called for FAILED events")
	}
	if len(reader.Committed) != 1 {
		t.Errorf("FAILED event offset should be committed (got %d commits)", len(reader.Committed))
	}
}

func TestStart_DBRetryAndRecovery(t *testing.T) {
	// DB fails first 2 attempts, succeeds on 3rd
	mockDB := &MockFailThenSucceedDB{FailCount: 2}
	reader := &MockReader{
		messages: []kafka.Message{
			{Value: []byte(`{"phoneNumber":"111","message":"hello","status":"SUCCESS"}`)},
		},
	}
	consumer := &Consumer{DB: mockDB, Reader: reader}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for len(reader.Committed) == 0 {
		}
		cancel()
	}()
	consumer.Start(ctx)

	if mockDB.calls != 3 {
		t.Errorf("Expected 3 SaveSMS calls (2 failures + 1 success), got %d", mockDB.calls)
	}
	if len(reader.Committed) != 1 {
		t.Errorf("Expected 1 committed message after recovery, got %d", len(reader.Committed))
	}
}
