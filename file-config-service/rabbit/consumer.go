package rabbit

import (
	"encoding/json"
	"file-config-service/models"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Conn      *amqp.Connection
	QueueName string
}

type PJobHandler func(string, string, string)

var JobHandler map[string]PJobHandler

func NewConsumer(queueName string) (*Consumer, error) {
	conn, err := amqp.Dial("amqp://rabbit:rabbit@192.168.1.10:5672")
	if err != nil {
		log.Println("Can't connect to RabbitMQ")
		return &Consumer{}, nil
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Println("Can't get channel")
		return &Consumer{}, nil
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
		return &Consumer{}, nil
	}

	JobHandler = make(map[string]PJobHandler)
	JobHandler["REGION"] = insertRegion
	JobHandler["TYPE"] = addType
	JobHandler["VERSION"] = addVersion
	return &Consumer{Conn: conn, QueueName: queueName}, nil
}

func (consumer *Consumer) Listen() error {

	channel, err := consumer.Conn.Channel()
	if err != nil {
		log.Println("Can't get channel when listening to queue")
		return err
	}
	defer channel.Close()

	messages, err := channel.Consume(
		consumer.QueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Can't consume messages")
		return err
	}

	go func() {
		for message := range messages {
			var payload Job
			_ = json.Unmarshal(message.Body, &payload)
			go JobHandler[payload.JobType](payload.Region, payload.Type, payload.Version)
		}
	}()

	wait := make(chan bool)
	<-wait
	return nil
}

func insertRegion(region, dbType, version string) {

	regionEntry := models.RegionEntry{
		Name:  region,
		Types: nil,
	}

	err := models.DB.RegionEntry.Insert(regionEntry)
	if err != nil {
		log.Println("Error when adding new region")
		return
	}
}

func addType(region, dbType, version string) {
	newType := models.Type{
		Name:     dbType,
		Versions: nil,
	}

	err := models.DB.RegionEntry.AddType(region, newType)
	if err != nil {
		log.Println("Error when adding new type")
		return
	}
}

func addVersion(region, dbType, version string) {
	newVersion := models.Version{
		Name: version,
	}

	err := models.DB.RegionEntry.AddVersion(region, dbType, newVersion)
	if err != nil {
		log.Println("Error when adding new version")
		return
	}
}
