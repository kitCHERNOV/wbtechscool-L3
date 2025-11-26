package models

type Notification struct {
	UUID   string
	Status string `json:"status"` // e.g., "pending", "sent", "failed"
	NotificationCard
}

type NotificationCard struct {
	UserID      string `json:"user_id"`
	Message     string `json:"message"`
	ScheduledAt int64  `json:"scheduled_at"`
}
