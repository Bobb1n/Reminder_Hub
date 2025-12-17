package rabbitmq

import (
	"testing"
	"time"

	"reminder-hub/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) PublishMessage(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *mockPublisher) IsPublished(msg interface{}) bool {
	args := m.Called(msg)
	return args.Bool(0)
}

func TestProducer_PublishEmail_Success(t *testing.T) {
	mockPub := new(mockPublisher)
	producer := &Producer{
		publisher: mockPub,
	}

	email := &models.RawEmail{
		EmailID:   "test-email-id",
		UserID:    "test-user-id",
		MessageID: "test-message-id",
		From:      "test@example.com",
		Subject:   "Test Subject",
		Text:      "Test Body",
		Date:      time.Now().Format(time.RFC3339),
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	mockPub.On("PublishMessage", email).Return(nil)

	err := producer.PublishEmail(email)

	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestProducer_PublishEmailBatch_EmptyBatch(t *testing.T) {
	mockPub := new(mockPublisher)
	producer := &Producer{
		publisher: mockPub,
	}

	emptyBatch := &models.RawEmails{
		RawEmail: []models.RawEmail{},
	}

	err := producer.PublishEmailBatch(emptyBatch)

	assert.NoError(t, err)
	mockPub.AssertNotCalled(t, "PublishMessage")
}

func TestProducer_PublishEmailBatch_Success(t *testing.T) {
	mockPub := new(mockPublisher)
	producer := &Producer{
		publisher: mockPub,
	}

	batch := &models.RawEmails{
		RawEmail: []models.RawEmail{
			{
				EmailID:   "email-1",
				UserID:    "user-1",
				MessageID: "msg-1",
				From:      "from1@example.com",
				Subject:   "Subject 1",
				Text:      "Body 1",
				Date:      "2025-12-17T10:00:00Z",
				TimeStamp: "2025-12-17T10:00:00Z",
			},
			{
				EmailID:   "email-2",
				UserID:    "user-2",
				MessageID: "msg-2",
				From:      "from2@example.com",
				Subject:   "Subject 2",
				Text:      "Body 2",
				Date:      "2025-12-17T11:00:00Z",
				TimeStamp: "2025-12-17T11:00:00Z",
			},
		},
	}

	mockPub.On("PublishMessage", mock.AnythingOfType("*models.RawEmails")).Return(nil)

	err := producer.PublishEmailBatch(batch)

	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestProducer_PublishEmailBatch_WithEmptyDates(t *testing.T) {
	mockPub := new(mockPublisher)
	producer := &Producer{
		publisher: mockPub,
	}

	batch := &models.RawEmails{
		RawEmail: []models.RawEmail{
			{
				EmailID:   "email-1",
				UserID:    "user-1",
				MessageID: "msg-1",
				From:      "from1@example.com",
				Subject:   "Subject 1",
				Text:      "Body 1",
				Date:      "", // Пустая дата
				TimeStamp: "", // Пустой timestamp
			},
		},
	}

	mockPub.On("PublishMessage", mock.AnythingOfType("*models.RawEmails")).Return(nil).Run(func(args mock.Arguments) {
		msg := args.Get(0).(*models.RawEmails)
		assert.Len(t, msg.RawEmail, 1)
		assert.NotEmpty(t, msg.RawEmail[0].Date)
		assert.NotEmpty(t, msg.RawEmail[0].TimeStamp)
	})

	err := producer.PublishEmailBatch(batch)

	assert.NoError(t, err)
	mockPub.AssertExpectations(t)
}

func TestProducer_Close(t *testing.T) {
	producer := &Producer{
		publisher: new(mockPublisher),
	}

	err := producer.Close()

	assert.NoError(t, err)
}

func TestNewProducerWithConn(t *testing.T) {

	assert.NotNil(t, NewProducerWithConn)
}
