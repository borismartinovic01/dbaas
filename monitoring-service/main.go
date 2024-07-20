package main

import (
	"bytes"
	"encoding/json"
	"log"
	"monitoring-service/rabbit"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var RabbitConnection *amqp.Connection
var Consumer *rabbit.Consumer

type DatasourceDto struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	URL       string `json:"url"`
	Access    string `json:"access"`
	IsDefault bool   `json:"isDefault"`
}

const (
	queueName string = "monitoring_queue"
)

func main() {

	registerPrometheusDatasource()

	RabbitConnection, err := rabbit.Connect()
	if err != nil {
		log.Println("Can't connect to RabbitMQ")
		os.Exit(1)
	}
	defer RabbitConnection.Close()

	Consumer, err = rabbit.NewConsumer(queueName, os.Getenv("TARGETS_FILE_PATH"), os.Getenv("POSTGRES_DASHBOARD_FILE_PATH"))
	if err != nil {
		log.Println("Can't create RabbitMQ Consumer")
		os.Exit(1)
	}
	go Consumer.Listen()

	wait := make(chan bool)
	<-wait
}

func registerPrometheusDatasource() {
	connCounts := 0
	maxCounts := 20

	payload := DatasourceDto{
		Name:      "Prometheus",
		Type:      "prometheus",
		URL:       "http://prometheus:9090",
		Access:    "proxy",
		IsDefault: true,
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(payload)
	if err != nil {
		log.Println("Can't encode datasource payload")
		return
	}

	url := "http://grafana:3000/api/datasources"
	request, err := http.NewRequest("POST", url, &body)
	if err != nil {
		log.Println("Error creating request to Grafana API")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth("admin", "admin")

	client := &http.Client{}
	for {

		if connCounts > maxCounts {
			panic("Can't configure Prometheus datasource in Grafana")
		}
		time.Sleep(10 * time.Second)

		response, err := client.Do(request)
		if err != nil {
			log.Println("Grafana not ready yet...")
			connCounts++
			continue
		}
		defer response.Body.Close()

		log.Println("Prometheus datasource configured successfully")
		return
	}
}
