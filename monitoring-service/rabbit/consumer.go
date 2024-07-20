package rabbit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Conn                      *amqp.Connection
	QueueName                 string
	TargetsMtx                sync.Mutex
	TargetsFilePath           string
	PostgresDashboardFilePath string
}

type Targets struct {
	Targets []string `json:"targets"`
}

func NewConsumer(queueName string, targetsFilePath string, postgresDashboardFilePath string) (*Consumer, error) {
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

	return &Consumer{
		Conn:                      conn,
		QueueName:                 queueName,
		TargetsFilePath:           targetsFilePath,
		PostgresDashboardFilePath: postgresDashboardFilePath}, nil
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
			go consumer.updatePrometheusTargets(payload.NodeIP, payload.NodePort)
			go consumer.createNewDashboard(payload.DashboardName, payload.DbType, fmt.Sprintf("%s:%s", payload.NodeIP, payload.NodePort), payload.Datname)
		}
	}()

	wait := make(chan bool)
	<-wait
	return nil
}

func (consumer *Consumer) updatePrometheusTargets(nodeIP, nodePort string) {
	consumer.TargetsMtx.Lock()
	defer consumer.TargetsMtx.Unlock()

	file, err := os.Open(consumer.TargetsFilePath)
	if err != nil {
		log.Println("Can't open prometheus targets file")
		return
	}
	defer file.Close()

	var targets []Targets
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&targets); err != nil {
		log.Println("Error while decoding prometheus targets file")
		return
	}

	targets[0].Targets = append(targets[0].Targets, fmt.Sprintf("%s:%s", nodeIP, nodePort))

	newTargets, _ := json.MarshalIndent(targets, "", " ")
	err = os.WriteFile(consumer.TargetsFilePath, newTargets, 0644)
	if err != nil {
		log.Println("Error writing new target to prometheus targets file")
		return
	}

	log.Println("Successfully wrote new target in prometheus targets file")
}

func (consumer *Consumer) createNewDashboard(dashboardName string, dbType string, instance string, datname string) {
	tempFilePath := dashboardName + ".json"

	switch dbType {
	case "PostgreSQL":
		templateData, err := os.ReadFile(consumer.PostgresDashboardFilePath)
		if err != nil {
			log.Println("Error reading PostgreSQL dashboard template")
			return
		}

		dashboardTemplate := string(templateData)
		dashboardTemplate = strings.ReplaceAll(dashboardTemplate, "$dashboardName", dashboardName)
		dashboardTemplate = strings.ReplaceAll(dashboardTemplate, "$instance", instance)
		dashboardTemplate = strings.ReplaceAll(dashboardTemplate, "$datname", datname)

		_ = os.WriteFile(tempFilePath, []byte(dashboardTemplate), 0644)
		tempData, _ := os.ReadFile(tempFilePath)

		var dashboardMap map[string]interface{}
		_ = json.Unmarshal(tempData, &dashboardMap)
		dashboardJSON, _ := json.Marshal(dashboardMap)
		_ = os.Remove(tempFilePath)

		url := "http://grafana:3000/api/dashboards/db"
		request, err := http.NewRequest("POST", url, bytes.NewBuffer(dashboardJSON))
		if err != nil {
			log.Println("Error creating request to Grafana API")
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.SetBasicAuth("admin", "admin")

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			log.Println("Error when creating dashboard through Grafana API")
			return
		}
		defer response.Body.Close()

		body, _ := io.ReadAll(response.Body)

		var responseData map[string]interface{}
		err = json.Unmarshal(body, &responseData)
		if err != nil {
			log.Println("Error getting Grafana response data")
			return
		}

		uid, _ := responseData["uid"].(string)
		consumer.sendUIDToConfigService(uid, dashboardName)
	}
}

func (consumer *Consumer) sendUIDToConfigService(grafanaUID string, deploymenUUID string) {

	requestPayload := SetGrafanaDto{
		GrafanaUID: grafanaUID,
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(requestPayload); err != nil {
		log.Println("Error when encoding payload for config service")
		return
	}

	url := "http://config-service:3002/databases/" + deploymenUUID
	request, err := http.NewRequest("PUT", url, &buffer)
	if err != nil {
		log.Println("Error creating request to Grafana API")
		return
	}
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth("admin", "admin")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Println("Error when creating dashboard through Grafana API")
		return
	}
	defer response.Body.Close()
}
