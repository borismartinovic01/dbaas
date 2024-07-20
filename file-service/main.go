package main

import (
	"file-service/rabbit"
	"file-service/utils"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitConnection *amqp.Connection
var Publisher *rabbit.Publisher

const (
	queueName string = "job_queue"
)

func main() {

	utils.InitUrl()

	RabbitConnection, err := rabbit.Connect()
	if err != nil {
		log.Println("Can't connect to RabbitMQ")
		os.Exit(1)
	}
	defer RabbitConnection.Close()

	Publisher, err = rabbit.NewPublisher(queueName)
	if err != nil {
		log.Println("Can't create RabbitMQ Publisher")
		os.Exit(1)
	}

	r := gin.Default()

	r.POST("/regions/:region/types/:type/versions/:version", UploadFile)
	r.GET("/regions/:region/types/:type/versions/:version", GetFile)
	r.POST("/regions", NewRegion)
	r.POST("/regions/:region/types", NewDatabaseType)
	r.POST("/regions/:region/types/:type/versions", NewVersion)

	r.Run(":3001")
}
