package main

import (
	"DelayedNotifier/internal/handlers"
	//"context"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SetBrokerConnection(connectionPath string) *amqp.Connection {
	conn, err := amqp.Dial(connectionPath)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	ch, _ := conn.Channel()
	defer ch.Close()

	// обменик отложенных сообщений
	args := amqp.Table{
		"x-delayed-type": "direct",
	}

	// Объявление контроллера отложенных сообщений
	const delayedExchange = "delayedExchange"
	err = ch.ExchangeDeclare(
		delayedExchange,
		"x-delayed-message", // type to delayed messages
		true,
		false,
		false,
		false,
		args,
	)

	if err != nil {
		log.Fatalf("Failed to declare delayed message exchanger: %s", err)
	}

	// основная очередь для передачи сообщений на обработку в consumer
	const workQueueName = "messageMainQueue"
	workQueue, err := ch.QueueDeclare(
		workQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a message queue: %s", err)
	}

	const myRoutingKey = "my_routing_key"
	_ = ch.QueueBind(
		workQueue.Name,
		myRoutingKey,
		delayedExchange,
		false,
		nil,
	)

	return conn
}

func main() {
	// config init
	//cfg :=

	// rabbitMQ init
	const brokerConnectionPath = "amqp://admin:password@localhost:5672/"
	conn := SetBrokerConnection(brokerConnectionPath)

	// Init Rest api
	// TODO: – POST /notify — создание уведомлений с датой и временем отправки;
	// TODO: – GET /notify/{id} — получение статуса уведомления;
	// TODO: – DELETE /notify/{id} — отмена запланированного уведомления.

	http.HandleFunc("/notify", handlers.NotificationRequest(conn))

	fmt.Println("Server is listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to launch http server: %s", err)
	}
}
