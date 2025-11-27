package handlers

import (
	"DelayedNotifier/internal/models"
	"DelayedNotifier/internal/rabbitMQ"
	"DelayedNotifier/internal/redisdb"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	//amqp "github.com/rabbitmq/amqp091-go"
)

// TODO: Описание основный http запросов

// TODO: Добавить реализацию для POST|GET|DELETE запросов по реквесту на один URI

const (
	qname = ""
)

// Post request to create notification
func CreateNotification(ctx context.Context, qp QueueProducer, rdb RedisStore, w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var notification models.Notification
	err = json.Unmarshal(data, &notification)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %s", err)
		http.Error(w, "Failed to unmarshal JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	// Добавление Id параметра в очередь
	err = qp.SendMessage(notification)
	if err != nil {
		log.Printf("Failed to send notification: %s", err)
		http.Error(w, "Failed to send notification: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Теперь сохранение в базу данных набора данных
	notification.Status = models.StatusPending
	err = rdb.SaveMessage(ctx, notification)
	if err != nil {
		log.Printf("Failed to save notification: %s", err)
		http.Error(w, "Failed to save notification", http.StatusInternalServerError)
		return
	}

	// TODO: to add some ResponseWriter parameters
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Message is created"))
	if err != nil {
		log.Printf("Failed to write response: %s", err)
		http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func GetNotificationStatus(ctx context.Context, rdb RedisStore, w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("id")

	status, err := rdb.GetStatus(ctx, uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"status": status}
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to marshal response: %s", err)
	}

	_, err = w.Write(jsonData)
	if err != nil {
		log.Printf("Failed to write response: %s", err)
	}

	return
}

func DeleteNotification(ctx context.Context, rdb RedisStore, w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("id")

	if uuid == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Notification deletion in progress"}`))

	go func(ctx context.Context) {
		err := rdb.DeleteMessage(ctx, uuid)
		if err != nil {
			log.Printf("Failed to delete message: %s", err)
		} else {
			log.Printf("Notification is deleted")
		}
	}(ctx)
}

// message string, timestamp int64
func NotificationRequest(ctx context.Context, conn *rabbitMQ.QueueProps, rdb *redisdb.RedisConnection) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CreateNotification(ctx, conn, rdb, w, r)
		case http.MethodGet:
			GetNotificationStatus(ctx, rdb, w, r)
		case http.MethodDelete:
			DeleteNotification(ctx, rdb, w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
