package models

const (
	StatusPending    = "pending"    // Ожидает отправки
	StatusProcessing = "processing" // В процессе отправки
	StatusSent       = "sent"       // Успешно отправлено
	StatusFailed     = "failed"     // Не удалось отправить
	StatusCancelled  = "cancelled"  // Отменено пользователем
	StatusRetrying   = "retrying"   // Повторная попытка отправки
)

type Notification struct {
	UUID   string `json:"uuid"`
	Status string `json:"status" redisdb:"status"` // e.g., "pending", "sent", "failed"
	NotificationCard
}

type NotificationCard struct {
	Message     string `json:"message" redisdb:"message"`
	ScheduledAt int64  `json:"scheduled_at" redisdb:"-"`
}
