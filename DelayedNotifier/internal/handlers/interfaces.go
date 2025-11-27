package handlers

import (
	"DelayedNotifier/internal/models"
	"context"
)

// QueueProducer interface for RabbitMQ operations
type QueueProducer interface {
	SendMessage(notification models.Notification) error
}

// RedisStore interface for Redis operations
type RedisStore interface {
	SaveMessage(ctx context.Context, notif models.Notification) error
	GetStatus(ctx context.Context, uuid string) (string, error)
	DeleteMessage(ctx context.Context, uuid string) error
}
