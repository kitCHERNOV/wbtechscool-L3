package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TestNotificationWithZeroDelay tests notification with 0ms delay
func TestNotificationWithZeroDelay(t *testing.T) {
	notifID := uuid.New().String()
	notification := Notification{
		UUID:        notifID,
		Message:     "Zero delay notification",
		ScheduledAt: 0,
	}

	body, _ := json.Marshal(notification)
	resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	t.Log("Zero delay notification created successfully")
	defer cleanupNotification(t, notifID)
}

// TestNotificationWithLongDelay tests notification with very long delay
func TestNotificationWithLongDelay(t *testing.T) {
	notifID := uuid.New().String()
	notification := Notification{
		UUID:        notifID,
		Message:     "Long delay notification",
		ScheduledAt: 3600000, // 1 hour
	}

	body, _ := json.Marshal(notification)
	resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	t.Log("Long delay notification created successfully")
	defer cleanupNotification(t, notifID)
}

// TestNotificationWithEmptyMessage tests notification with empty message
func TestNotificationWithEmptyMessage(t *testing.T) {
	notifID := uuid.New().String()
	notification := Notification{
		UUID:        notifID,
		Message:     "",
		ScheduledAt: shortDelayMs,
	}

	body, _ := json.Marshal(notification)
	resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	defer resp.Body.Close()

	// Current implementation doesn't validate empty message
	// This test documents the behavior
	if resp.StatusCode == http.StatusCreated {
		t.Log("Empty message notification was accepted (no validation)")
		defer cleanupNotification(t, notifID)
	} else {
		t.Logf("Empty message notification rejected with status: %d", resp.StatusCode)
	}
}

// TestNotificationWithVeryLongMessage tests notification with large message
func TestNotificationWithVeryLongMessage(t *testing.T) {
	notifID := uuid.New().String()
	longMessage := string(make([]byte, 10000))
	for i := range longMessage {
		longMessage = longMessage[:i] + "A" + longMessage[i+1:]
	}

	notification := Notification{
		UUID:        notifID,
		Message:     longMessage,
		ScheduledAt: shortDelayMs,
	}

	body, _ := json.Marshal(notification)
	resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		t.Log("Very long message notification created successfully")
		defer cleanupNotification(t, notifID)
	} else {
		t.Logf("Very long message rejected with status: %d", resp.StatusCode)
	}
}

// TestNotificationWithSpecialCharacters tests notification with special characters
func TestNotificationWithSpecialCharacters(t *testing.T) {
	specialMessages := []string{
		"Hello 世界",
		"Test with emoji 🚀🎉",
		"Special chars: <>&\"'",
		"Newline\nand\ttab",
		"JSON breaking: {\"key\": \"value\"}",
	}

	for i, msg := range specialMessages {
		t.Run(msg[:min(10, len(msg))], func(t *testing.T) {
			notifID := uuid.New().String()
			notification := Notification{
				UUID:        notifID,
				Message:     msg,
				ScheduledAt: shortDelayMs,
			}

			body, _ := json.Marshal(notification)
			resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create notification %d: %v", i, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				t.Errorf("Special chars message %d rejected with status: %d", i, resp.StatusCode)
			} else {
				t.Logf("Special chars message %d accepted", i)
				defer cleanupNotification(t, notifID)
			}
		})
	}
}

// TestNotificationWithMissingFields tests notifications with missing required fields
func TestNotificationWithMissingFields(t *testing.T) {
	testCases := []struct {
		name string
		body map[string]interface{}
	}{
		{
			name: "Missing UUID",
			body: map[string]interface{}{
				"message":      "Test message",
				"scheduled_at": shortDelayMs,
			},
		},
		{
			name: "Missing Message",
			body: map[string]interface{}{
				"uuid":         uuid.New().String(),
				"scheduled_at": shortDelayMs,
			},
		},
		{
			name: "Missing ScheduledAt",
			body: map[string]interface{}{
				"uuid":    uuid.New().String(),
				"message": "Test message",
			},
		},
		{
			name: "Empty Object",
			body: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.body)
			resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			// Current implementation may accept these - this documents the behavior
			t.Logf("%s: status %d", tc.name, resp.StatusCode)

			// Cleanup if created
			if resp.StatusCode == http.StatusCreated {
				if uuidVal, ok := tc.body["uuid"].(string); ok {
					defer cleanupNotification(t, uuidVal)
				}
			}
		})
	}
}

// TestDuplicateUUID tests creating notifications with duplicate UUIDs
func TestDuplicateUUID(t *testing.T) {
	notifID := uuid.New().String()
	notification := Notification{
		UUID:        notifID,
		Message:     "First notification",
		ScheduledAt: shortDelayMs,
	}

	// Create first notification
	body, _ := json.Marshal(notification)
	resp1, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create first notification: %v", err)
	}
	resp1.Body.Close()

	if resp1.StatusCode != http.StatusCreated {
		t.Fatalf("First notification creation failed with status: %d", resp1.StatusCode)
	}

	// Try to create duplicate
	notification.Message = "Second notification with same UUID"
	body, _ = json.Marshal(notification)
	resp2, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to send duplicate notification: %v", err)
	}
	defer resp2.Body.Close()

	// Current implementation may allow duplicates - this documents the behavior
	t.Logf("Duplicate UUID notification status: %d", resp2.StatusCode)

	defer cleanupNotification(t, notifID)
}

// TestDeleteNonExistentNotification tests deleting a notification that doesn't exist
func TestDeleteNonExistentNotification(t *testing.T) {
	nonExistentID := uuid.New().String()
	deleteURL := baseURL + "/notify/" + nonExistentID

	req, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	// Current implementation accepts delete requests asynchronously
	// It should return 202 even for non-existent IDs
	if resp.StatusCode != http.StatusAccepted {
		t.Logf("Non-existent notification delete status: %d", resp.StatusCode)
	} else {
		t.Log("Delete request accepted for non-existent notification")
	}
}

// TestRedisConnectionFailureRecovery tests behavior when Redis is temporarily unavailable
func TestRedisDataIntegrity(t *testing.T) {
	ctx := context.Background()
	notifID := uuid.New().String()
	testMessage := "Data integrity test"

	// Create notification
	notification := Notification{
		UUID:        notifID,
		Message:     testMessage,
		ScheduledAt: shortDelayMs,
	}

	body, _ := json.Marshal(notification)
	resp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create notification: %v", err)
	}
	resp.Body.Close()

	// Wait for data to be saved
	time.Sleep(500 * time.Millisecond)

	// Verify data in Redis directly
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})
	defer rdb.Close()

	// Check all fields
	exists, err := rdb.Exists(ctx, notifID).Result()
	if err != nil {
		t.Errorf("Failed to check existence: %v", err)
	}
	if exists == 0 {
		t.Error("Notification not found in Redis")
	}

	message, err := rdb.HGet(ctx, notifID, "message").Result()
	if err != nil {
		t.Errorf("Failed to get message: %v", err)
	}
	if message != testMessage {
		t.Errorf("Message mismatch: expected '%s', got '%s'", testMessage, message)
	}

	status, err := rdb.HGet(ctx, notifID, "status").Result()
	if err != nil {
		t.Errorf("Failed to get status: %v", err)
	}
	if status != "pending" {
		t.Errorf("Status mismatch: expected 'pending', got '%s'", status)
	}

	t.Log("Redis data integrity verified")
	defer cleanupNotification(t, notifID)
}

// TestRapidCreateAndDelete tests quickly creating and deleting notifications
func TestRapidCreateAndDelete(t *testing.T) {
	const iterations = 5
	for i := 0; i < iterations; i++ {
		notifID := uuid.New().String()
		notification := Notification{
			UUID:        notifID,
			Message:     "Rapid test notification",
			ScheduledAt: mediumDelayMs,
		}

		// Create
		body, _ := json.Marshal(notification)
		createResp, err := http.Post(baseURL+"/notify", "application/json", bytes.NewBuffer(body))
		if err != nil {
			t.Errorf("Iteration %d: Failed to create: %v", i, err)
			continue
		}
		createResp.Body.Close()

		// Immediately delete
		deleteURL := baseURL + "/notify/" + notifID
		deleteReq, _ := http.NewRequest(http.MethodDelete, deleteURL, nil)
		deleteResp, err := http.DefaultClient.Do(deleteReq)
		if err != nil {
			t.Errorf("Iteration %d: Failed to delete: %v", i, err)
		} else {
			deleteResp.Body.Close()
		}

		// Small delay between iterations
		time.Sleep(100 * time.Millisecond)
	}

	t.Log("Rapid create and delete test completed")
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
