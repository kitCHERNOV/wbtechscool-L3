package handlers

import (
	"DelayedNotifier/internal/models"
	"DelayedNotifier/internal/rabbitMQ"
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
func CreateNotification(qp *rabbitMQ.QueueProps, w http.ResponseWriter, r *http.Request) http.ResponseWriter {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %s", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return w
	}

	var notification models.Notification
	err = json.Unmarshal(data, &notification)
	if err != nil {
		log.Printf("Failed to unmarshal JSON: %s", err)
		http.Error(w, "Failed to unmarshal JSON: "+err.Error(), http.StatusBadRequest)
		return w
	}

	err = qp.SendMessage(notification)
	if err != nil {
		log.Printf("Failed to send notification: %s", err)
		http.Error(w, "Failed to send notification: "+err.Error(), http.StatusInternalServerError)
		return w
	}

	// TODO: to add some ResponseWriter parameters
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("Message is created"))
	if err != nil {
		log.Printf("Failed to write response: %s", err)
		http.Error(w, "Failed to write response: "+err.Error(), http.StatusInternalServerError)
		return w
	}
	return w
}

// message string, timestamp int64
func NotificationRequest(conn *rabbitMQ.QueueProps) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w = CreateNotification(conn, w, r)

		}
		if r.Method == http.MethodGet {

		}
		if r.Method == http.MethodDelete {

		}
	}
}
