package handlers

import (
	"DelayedNotifier/internal/models"
	"encoding/json"
	"io"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO: Описание основный http запросов

// TODO: Добавить реализацию для POST|GET|DELETE запросов по реквесту на один URI

// Post request to create notification
func CreateNotification(conn *amqp.Connection, w http.ResponseWriter, r *http.Request) {
	var notification models.Notification
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		return
	}
	err = json.Unmarshal(data, &notification)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %s", err)
		return
	}

}

// message string, timestamp int64
func NotificationRequest(conn *amqp.Connection) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {

		}
	}

}
