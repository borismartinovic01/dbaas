package rabbit

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect() (*amqp.Connection, error) {
	connCounts := 0
	maxCounts := 20

	for {
		conn, err := amqp.Dial("amqp://rabbit:rabbit@192.168.1.10:5672")
		if err == nil {
			log.Println("Connected to RabbitMQ")
			return conn, nil
		}

		log.Println(err)
		log.Println("RabbitMQ not ready yet...")
		connCounts++

		if connCounts > maxCounts {
			return nil, err
		}

		time.Sleep(2 * time.Second)
		continue
	}
}
