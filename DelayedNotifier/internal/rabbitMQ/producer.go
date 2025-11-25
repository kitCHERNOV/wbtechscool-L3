package rabbitMQ

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueProps struct {
	Channel *amqp.Channel
}

func NewQueueProps(ch *amqp.Channel) *QueueProps {
	return &QueueProps{Channel: ch}
}

// CreateMainQueue declares a queue to hold messages.
func (qp *QueueProps) CreateMainQueue(queueName string) (amqp.Queue, error) {
	q, err := qp.Channel.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Printf("Failed to declare a queue: %s", err)
		return amqp.Queue{}, err
	}
	return q, nil
}

// SendMessage publishes a message to the specified queue.
func (qp *QueueProps) SendMessage(queueName string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := qp.Channel.PublishWithContext(ctx,
		"",        // exchange (default)
		queueName, // routing key (the queue name)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		log.Printf("Failed to publish a message: %s", err)
		return err
	}
	log.Printf(" [x] Sent %s to queue %s\n", body, queueName)
	return nil
}
