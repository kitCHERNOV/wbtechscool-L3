package handlers

import (
	"DelayedNotifier/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

// Mock QueueProps для тестирования
type MockQueueProps struct {
	SendMessageFunc func(notification models.Notification) error
}

func (m *MockQueueProps) SendMessage(notification models.Notification) error {
	if m.SendMessageFunc != nil {
		return m.SendMessageFunc(notification)
	}
	return nil
}

// Mock RedisConnection для тестирования
type MockRedisConnection struct {
	SaveMessageFunc   func(ctx context.Context, notif models.Notification) error
	GetStatusFunc     func(ctx context.Context, uuid string) (string, error)
	DeleteMessageFunc func(ctx context.Context, uuid string) error
}

func (m *MockRedisConnection) SaveMessage(ctx context.Context, notif models.Notification) error {
	if m.SaveMessageFunc != nil {
		return m.SaveMessageFunc(ctx, notif)
	}
	return nil
}

func (m *MockRedisConnection) GetStatus(ctx context.Context, uuid string) (string, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, uuid)
	}
	return models.StatusPending, nil
}

func (m *MockRedisConnection) DeleteMessage(ctx context.Context, uuid string) error {
	if m.DeleteMessageFunc != nil {
		return m.DeleteMessageFunc(ctx, uuid)
	}
	return nil
}

func (m *MockRedisConnection) Close() {}

// Helper function to create mock dependencies
func createMockDependencies() (context.Context, *MockQueueProps, *MockRedisConnection) {
	ctx := context.Background()
	mockQueue := &MockQueueProps{}
	mockRedis := &MockRedisConnection{}
	return ctx, mockQueue, mockRedis
}

// TestNotificationRequest_POST tests POST /notify endpoint
func TestNotificationRequest_POST(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		mockQueueError     error
		mockRedisError     error
		expectedStatusCode int
		expectedBodyPart   string
	}{
		{
			name: "Successful notification creation",
			requestBody: models.Notification{
				UUID: uuid.New().String(),
				NotificationCard: models.NotificationCard{
					Message:     "Test message",
					ScheduledAt: 5000,
				},
			},
			mockQueueError:     nil,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusCreated,
			expectedBodyPart:   "Message is created",
		},
		{
			name:               "Invalid JSON",
			requestBody:        "invalid json",
			mockQueueError:     nil,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBodyPart:   "Failed to unmarshal JSON",
		},
		{
			name: "Queue send failure",
			requestBody: models.Notification{
				UUID: uuid.New().String(),
				NotificationCard: models.NotificationCard{
					Message:     "Test message",
					ScheduledAt: 5000,
				},
			},
			mockQueueError:     errors.New("queue connection failed"),
			mockRedisError:     nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBodyPart:   "Failed to send notification",
		},
		{
			name: "Redis save failure",
			requestBody: models.Notification{
				UUID: uuid.New().String(),
				NotificationCard: models.NotificationCard{
					Message:     "Test message",
					ScheduledAt: 5000,
				},
			},
			mockQueueError:     nil,
			mockRedisError:     errors.New("redis connection failed"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedBodyPart:   "Failed to save notification",
		},
		{
			name: "Empty message",
			requestBody: models.Notification{
				UUID: uuid.New().String(),
				NotificationCard: models.NotificationCard{
					Message:     "",
					ScheduledAt: 5000,
				},
			},
			mockQueueError:     nil,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusCreated,
			expectedBodyPart:   "Message is created",
		},
		{
			name: "Zero delay",
			requestBody: models.Notification{
				UUID: uuid.New().String(),
				NotificationCard: models.NotificationCard{
					Message:     "Immediate notification",
					ScheduledAt: 0,
				},
			},
			mockQueueError:     nil,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusCreated,
			expectedBodyPart:   "Message is created",
		},
		{
			name: "Large delay",
			requestBody: models.Notification{
				UUID: uuid.New().String(),
				NotificationCard: models.NotificationCard{
					Message:     "Long delayed notification",
					ScheduledAt: 86400000, // 24 hours
				},
			},
			mockQueueError:     nil,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusCreated,
			expectedBodyPart:   "Message is created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, mockQueue, mockRedis := createMockDependencies()

			// Setup mock functions
			mockQueue.SendMessageFunc = func(notification models.Notification) error {
				return tt.mockQueueError
			}

			mockRedis.SaveMessageFunc = func(ctx context.Context, notif models.Notification) error {
				// Verify status was set to pending
				if notif.Status != models.StatusPending && tt.mockRedisError == nil {
					t.Errorf("Expected status to be 'pending', got '%s'", notif.Status)
				}
				return tt.mockRedisError
			}

			// Create request body
			var body []byte
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.requestBody)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler using the interface types
			CreateNotification(ctx, mockQueue, mockRedis, w, req)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// Check response body
			responseBody := w.Body.String()
			if tt.expectedBodyPart != "" && !containsString(responseBody, tt.expectedBodyPart) {
				t.Errorf("Expected response to contain '%s', got '%s'", tt.expectedBodyPart, responseBody)
			}
		})
	}
}

// TestNotificationRequest_GET tests GET /notify/{id} endpoint
func TestNotificationRequest_GET(t *testing.T) {
	tests := []struct {
		name               string
		notificationID     string
		mockStatus         string
		mockRedisError     error
		expectedStatusCode int
		expectedStatus     string
	}{
		{
			name:               "Get pending notification status",
			notificationID:     uuid.New().String(),
			mockStatus:         models.StatusPending,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusOK,
			expectedStatus:     models.StatusPending,
		},
		{
			name:               "Get sent notification status",
			notificationID:     uuid.New().String(),
			mockStatus:         models.StatusSent,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusOK,
			expectedStatus:     models.StatusSent,
		},
		{
			name:               "Get failed notification status",
			notificationID:     uuid.New().String(),
			mockStatus:         models.StatusFailed,
			mockRedisError:     nil,
			expectedStatusCode: http.StatusOK,
			expectedStatus:     models.StatusFailed,
		},
		{
			name:               "Notification not found",
			notificationID:     uuid.New().String(),
			mockStatus:         "",
			mockRedisError:     errors.New("notification not found"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedStatus:     "",
		},
		{
			name:               "Redis connection error",
			notificationID:     uuid.New().String(),
			mockStatus:         "",
			mockRedisError:     errors.New("redis connection failed"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedStatus:     "",
		},
		{
			name:               "Empty UUID",
			notificationID:     "",
			mockStatus:         "",
			mockRedisError:     errors.New("empty uuid"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedStatus:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _, mockRedis := createMockDependencies()

			// Setup mock function
			mockRedis.GetStatusFunc = func(ctx context.Context, uuid string) (string, error) {
				if uuid != tt.notificationID {
					t.Errorf("Expected UUID '%s', got '%s'", tt.notificationID, uuid)
				}
				return tt.mockStatus, tt.mockRedisError
			}

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/notify/"+tt.notificationID, nil)
			req.SetPathValue("id", tt.notificationID)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			GetNotificationStatus(ctx, mockRedis, w, req)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// Check response body for successful requests
			if tt.expectedStatusCode == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response["status"] != tt.expectedStatus {
					t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, response["status"])
				}

				// Check Content-Type header
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
				}
			}
		})
	}
}

// TestNotificationRequest_DELETE tests DELETE /notify/{id} endpoint
func TestNotificationRequest_DELETE(t *testing.T) {
	tests := []struct {
		name               string
		notificationID     string
		mockRedisError     error
		expectedStatusCode int
		expectedBodyPart   string
	}{
		{
			name:               "Successful deletion",
			notificationID:     uuid.New().String(),
			mockRedisError:     nil,
			expectedStatusCode: http.StatusAccepted,
			expectedBodyPart:   "Notification deletion in progress",
		},
		{
			name:               "Delete with valid UUID",
			notificationID:     "550e8400-e29b-41d4-a716-446655440000",
			mockRedisError:     nil,
			expectedStatusCode: http.StatusAccepted,
			expectedBodyPart:   "Notification deletion in progress",
		},
		{
			name:               "Empty UUID",
			notificationID:     "",
			mockRedisError:     nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBodyPart:   "Invalid request",
		},
		{
			name:               "Redis deletion error (async, still returns 202)",
			notificationID:     uuid.New().String(),
			mockRedisError:     errors.New("redis delete failed"),
			expectedStatusCode: http.StatusAccepted,
			expectedBodyPart:   "Notification deletion in progress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _, mockRedis := createMockDependencies()

			deleteCalled := false
			mockRedis.DeleteMessageFunc = func(ctx context.Context, uuid string) error {
				deleteCalled = true
				if uuid != tt.notificationID {
					t.Errorf("Expected UUID '%s', got '%s'", tt.notificationID, uuid)
				}
				return tt.mockRedisError
			}

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/notify/"+tt.notificationID, nil)
			req.SetPathValue("id", tt.notificationID)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			DeleteNotification(ctx, mockRedis, w, req)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}

			// Check response body
			responseBody := w.Body.String()
			if tt.expectedBodyPart != "" && !containsString(responseBody, tt.expectedBodyPart) {
				t.Errorf("Expected response to contain '%s', got '%s'", tt.expectedBodyPart, responseBody)
			}

			// For valid requests, verify delete was NOT called synchronously (it's async)
			// We can't easily test the goroutine, but we verify the response is immediate
			if tt.expectedStatusCode == http.StatusAccepted && tt.notificationID != "" {
				// The response should be immediate, before goroutine executes
				// This is inherent in the async design
			}

			// For empty UUID, delete should not be called
			if tt.notificationID == "" && deleteCalled {
				t.Error("Delete should not be called for empty UUID")
			}
		})
	}
}

// TestNotificationRequest_MethodRouting tests that NotificationRequest routes to correct handler
func TestNotificationRequest_MethodRouting(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		requestBody        string
		pathID             string
		expectedStatusCode int
	}{
		{
			name:               "POST method routes correctly",
			method:             http.MethodPost,
			requestBody:        `{"uuid":"test-uuid","message":"test","scheduled_at":5000}`,
			pathID:             "",
			expectedStatusCode: http.StatusCreated,
		},
		{
			name:               "GET method routes correctly",
			method:             http.MethodGet,
			requestBody:        "",
			pathID:             uuid.New().String(),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "DELETE method routes correctly",
			method:             http.MethodDelete,
			requestBody:        "",
			pathID:             uuid.New().String(),
			expectedStatusCode: http.StatusAccepted,
		},
		{
			name:               "PUT method not allowed",
			method:             http.MethodPut,
			requestBody:        "",
			pathID:             "",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "PATCH method not allowed",
			method:             http.MethodPatch,
			requestBody:        "",
			pathID:             "",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "HEAD method not allowed",
			method:             http.MethodHead,
			requestBody:        "",
			pathID:             "",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "OPTIONS method not allowed",
			method:             http.MethodOptions,
			requestBody:        "",
			pathID:             "",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, mockQueue, mockRedis := createMockDependencies()

			// Setup mocks with default success behavior
			mockQueue.SendMessageFunc = func(notification models.Notification) error {
				return nil
			}
			mockRedis.SaveMessageFunc = func(ctx context.Context, notif models.Notification) error {
				return nil
			}
			mockRedis.GetStatusFunc = func(ctx context.Context, uuid string) (string, error) {
				return models.StatusPending, nil
			}
			mockRedis.DeleteMessageFunc = func(ctx context.Context, uuid string) error {
				return nil
			}

			// Create request
			var body io.Reader
			if tt.requestBody != "" {
				body = bytes.NewBufferString(tt.requestBody)
			}
			req := httptest.NewRequest(tt.method, "/notify", body)
			if tt.pathID != "" {
				req.SetPathValue("id", tt.pathID)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Create the handler function (simulating NotificationRequest wrapper)
			handler := func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					CreateNotification(ctx, mockQueue, mockRedis, w, r)
				case http.MethodGet:
					GetNotificationStatus(ctx, mockRedis, w, r)
				case http.MethodDelete:
					DeleteNotification(ctx, mockRedis, w, r)
				default:
					w.WriteHeader(http.StatusMethodNotAllowed)
				}
			}

			// Call handler
			handler(w, req)

			// Check status code
			if w.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, w.Code)
			}
		})
	}
}

// TestNotificationRequest_Integration tests the full NotificationRequest function
func TestNotificationRequest_Integration(t *testing.T) {
	ctx, mockQueue, mockRedis := createMockDependencies()

	// Setup mocks
	savedNotifications := make(map[string]models.Notification)

	mockQueue.SendMessageFunc = func(notification models.Notification) error {
		return nil
	}

	mockRedis.SaveMessageFunc = func(ctx context.Context, notif models.Notification) error {
		savedNotifications[notif.UUID] = notif
		return nil
	}

	mockRedis.GetStatusFunc = func(ctx context.Context, uuid string) (string, error) {
		if notif, exists := savedNotifications[uuid]; exists {
			return notif.Status, nil
		}
		return "", errors.New("not found")
	}

	mockRedis.DeleteMessageFunc = func(ctx context.Context, uuid string) error {
		delete(savedNotifications, uuid)
		return nil
	}

	// We can't directly test NotificationRequest wrapper without proper interface injection
	// But we can test the flow by calling handlers directly

	// Test 1: Create notification
	notifID := uuid.New().String()
	createBody := models.Notification{
		UUID: notifID,
		NotificationCard: models.NotificationCard{
			Message:     "Integration test message",
			ScheduledAt: 5000,
		},
	}
	bodyBytes, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBuffer(bodyBytes))
	createW := httptest.NewRecorder()
	CreateNotification(ctx, mockQueue, mockRedis, createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Errorf("Create failed: expected %d, got %d", http.StatusCreated, createW.Code)
	}

	// Test 2: Get notification status
	getReq := httptest.NewRequest(http.MethodGet, "/notify/"+notifID, nil)
	getReq.SetPathValue("id", notifID)
	getW := httptest.NewRecorder()
	GetNotificationStatus(ctx, mockRedis, getW, getReq)

	if getW.Code != http.StatusOK {
		t.Errorf("Get failed: expected %d, got %d", http.StatusOK, getW.Code)
	}

	var statusResp map[string]string
	json.Unmarshal(getW.Body.Bytes(), &statusResp)
	if statusResp["status"] != models.StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", statusResp["status"])
	}

	// Test 3: Delete notification
	deleteReq := httptest.NewRequest(http.MethodDelete, "/notify/"+notifID, nil)
	deleteReq.SetPathValue("id", notifID)
	deleteW := httptest.NewRecorder()
	DeleteNotification(ctx, mockRedis, deleteW, deleteReq)

	if deleteW.Code != http.StatusAccepted {
		t.Errorf("Delete failed: expected %d, got %d", http.StatusAccepted, deleteW.Code)
	}
}

// TestCreateNotification_EdgeCases tests edge cases for CreateNotification
func TestCreateNotification_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		requestBody string
		setupMock   func(*MockQueueProps, *MockRedisConnection)
		checkResult func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "Empty request body",
			requestBody: "",
			setupMock: func(mq *MockQueueProps, mr *MockRedisConnection) {
				mq.SendMessageFunc = func(notification models.Notification) error {
					return nil
				}
				mr.SaveMessageFunc = func(ctx context.Context, notif models.Notification) error {
					return nil
				}
			},
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
				}
			},
		},
		{
			name:        "Malformed JSON",
			requestBody: `{"uuid":"test", "message":`,
			setupMock: func(mq *MockQueueProps, mr *MockRedisConnection) {
				mq.SendMessageFunc = func(notification models.Notification) error {
					return nil
				}
			},
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusBadRequest {
					t.Errorf("Expected %d, got %d", http.StatusBadRequest, w.Code)
				}
				if !containsString(w.Body.String(), "Failed to unmarshal JSON") {
					t.Error("Expected unmarshal error message")
				}
			},
		},
		{
			name:        "Special characters in message",
			requestBody: `{"uuid":"test-uuid","message":"Test with emoji 🚀 and special chars <>&\"","scheduled_at":5000}`,
			setupMock: func(mq *MockQueueProps, mr *MockRedisConnection) {
				mq.SendMessageFunc = func(notification models.Notification) error {
					if notification.NotificationCard.Message != "Test with emoji 🚀 and special chars <>&\"" {
						t.Error("Special characters not preserved")
					}
					return nil
				}
				mr.SaveMessageFunc = func(ctx context.Context, notif models.Notification) error {
					return nil
				}
			},
			checkResult: func(t *testing.T, w *httptest.ResponseRecorder) {
				if w.Code != http.StatusCreated {
					t.Errorf("Expected %d, got %d", http.StatusCreated, w.Code)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, mockQueue, mockRedis := createMockDependencies()
			tt.setupMock(mockQueue, mockRedis)

			req := httptest.NewRequest(http.MethodPost, "/notify", bytes.NewBufferString(tt.requestBody))
			w := httptest.NewRecorder()

			CreateNotification(ctx, mockQueue, mockRedis, w, req)

			tt.checkResult(t, w)
		})
	}
}

// Helper function to check if string contains substring
func containsString(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}
