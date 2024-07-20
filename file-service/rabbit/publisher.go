package rabbit

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	Conn      *amqp.Connection
	QueueName string
}

func NewPublisher(queueName string) (*Publisher, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Println("Can't connect to RabbitMQ")
		return &Publisher{}, nil
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Println("Can't get channel")
		return &Publisher{}, nil
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		queueName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Can't create queue")
		return &Publisher{}, nil
	}

	return &Publisher{Conn: conn, QueueName: queueName}, nil
}

func (publisher *Publisher) Push(body []byte) error {

	channel, err := publisher.Conn.Channel()
	if err != nil {
		log.Println("Can't get channel when pushing to queue")
		return err
	}
	defer channel.Close()

	message := amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	}

	err = channel.Publish(
		"",
		publisher.QueueName,
		false,
		false,
		message,
	)
	if err != nil {
		log.Println("Can't publish the message")
		return err
	}

	return nil
}
