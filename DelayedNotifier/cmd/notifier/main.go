package main

import (
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/handlers"
	"DelayedNotifier/internal/rabbitMQ"
	"DelayedNotifier/internal/redisdb"
	"context"
	"net"

	//"context"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

func SetBrokerConnection(connectionPath string) *amqp.Channel {
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

	return ch
}

func main() {
	// config init
	cfg := config.MustLoad()

	// Radis connection init
	rdb := redisdb.DeclareRedisDataBase(redis.Options{
		Addr:     net.JoinHostPort(cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
		Protocol: 2,
	})
	defer rdb.Close()

	// rabbitMQ init
	const brokerConnectionPath = "amqp://admin:password@localhost:5672/"
	conn, err := amqp.Dial(brokerConnectionPath)
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
	// create producer's storage
	channel := rabbitMQ.NewQueueProps(ch, delayedExchange, myRoutingKey)

	// Init Rest api
	// TODO: – POST /notify — создание уведомлений с датой и временем отправки;
	// TODO: – GET /notify/{id} — получение статуса уведомления;
	// TODO: – DELETE /notify/{id} — отмена запланированного уведомления.

	ctx := context.Background()

	http.HandleFunc("/notify", handlers.NotificationRequest(ctx, channel, rdb))

	fmt.Println("Server is listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to launch http server: %s", err)
	}
}
