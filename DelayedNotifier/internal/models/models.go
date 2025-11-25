package models

import "time"

type Notification struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"` // Unix timestamp
	Status    string `json:"status"`    // e.g., "pending", "sent", "failed"
}

type CreateNotificationRequest struct {
	UserID      string    `json:"user_id"`
	Message     string    `json:"message"`
	ScheduledAt time.Time `json:"scheduled_at"`
}
