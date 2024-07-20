package main

import (
	"context"
	controller "file-config-service/controllers"
	"file-config-service/models"
	"file-config-service/rabbit"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var mongoUrl string
var dbName string
var dbUsername string
var dbPassword string

var RabbitConnection *amqp.Connection
var Consumer *rabbit.Consumer

const (
	queueName string = "job_queue"
)

func main() {
	mongoUrl = os.Getenv("DSN")
	dbName = os.Getenv("DB_NAME")
	dbUsername = os.Getenv("DB_USER")
	dbPassword = os.Getenv("DB_PASSWORD")

	c, err := createMongoClient()
	if err != nil {
		log.Panic(err)
	}
	client = c

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	models.New(dbName, client)

	RabbitConnection, err := rabbit.Connect()
	if err != nil {
		log.Println("Can't connect to RabbitMQ")
		os.Exit(1)
	}
	defer RabbitConnection.Close()

	Consumer, err = rabbit.NewConsumer(queueName)
	if err != nil {
		log.Println("Can't create RabbitMQ Consumer")
		os.Exit(1)
	}
	go Consumer.Listen()

	r := gin.Default()

	r.GET("/regions", controller.GetAll)

	r.Run()
}

func createMongoClient() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoUrl)
	clientOptions.SetAuth(options.Credential{
		Username: dbUsername,
		Password: dbPassword,
	})

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println("Error when connecting to mongo: ", err)
		return nil, err
	}

	return client, nil
}
