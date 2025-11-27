package rabbitMQ

import (
	"DelayedNotifier/internal/models"
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueProps struct {
	Channel         *amqp.Channel
	WaitingExchange string
	RoutingKey      string
	// Add Redis DataBase... to storage
}

func NewQueueProps(ch *amqp.Channel, args ...string) *QueueProps {
	return &QueueProps{
		Channel:         ch,
		WaitingExchange: args[0],
		RoutingKey:      args[1],
	}
}

// SendMessage publishes a message to the specified queue.
func (qp *QueueProps) SendMessage(notification models.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	//millisecondsDelay := time.Duration(notification.ScheduledAt) * time.Millisecond
	headers := amqp.Table{
		"x-delay": int64(notification.ScheduledAt),
	}
	// Хранить в очереди будем только message UUID поле.
	err := qp.Channel.PublishWithContext(
		ctx,
		qp.WaitingExchange,
		qp.RoutingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(notification.UUID),
			Headers:      headers,
		},
	)

	if err != nil {
		log.Printf("Failed to publish a message: %s", err)
		return err
	}

	return nil
}

func (qp *QueueProps) GetMessageStatus(messageId string) (models.Notification, error) {
	// TODO: to implement get status of certain message

	return models.Notification{}, nil
}

func (qp *QueueProps) CancelMessageDelay(messageId string) (models.Notification, error) {
	return models.Notification{}, nil
}
