package main

import (
	"config-service/controllers"
	"config-service/models"
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var mongoUrl string
var dbName string
var dbUsername string
var dbPassword string
var LocationServerMp map[string]string

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

	router := gin.Default()

	router.POST("/servers", controllers.CreateServer)
	router.POST("/databases", controllers.CreateDatabase)
	router.GET("/users/:email/databases/:name", controllers.DbOverviewByName)
	router.GET("/users/:email/databases", controllers.UserDatabases)
	router.GET("/users/:email/servers", controllers.UserServers)
	router.GET("/users/:email/databases/:name/grafana", controllers.DbGrafanaUIDByName)
	router.PUT("/databases/:directoryUUID", controllers.UpdateGrafanaUIDByDirectoryUUID)

	router.Run()
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
