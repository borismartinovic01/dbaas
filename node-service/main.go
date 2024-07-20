package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"node-service/rabbit"
	"node-service/utils"
	"os"
	"sync"

	"github.com/go-redis/redis"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RPCServer struct {
	portMtx      sync.Mutex
	portReserved map[int]int
	redisClient  *redis.Client
}

type ConnectPayload struct {
	ClientIp string
}

type App struct {
	MyIP string
}

var app *App

var RabbitConnection *amqp.Connection
var Publisher *rabbit.Publisher

const (
	queueName string = "monitoring_queue"
)

func main() {

	utils.InitUrl()

	rpc.Register(NewRPCServer())
	rpc.HandleHTTP()
	go app.listenRPC()

	app = &App{
		MyIP: "192.168.1.11:3000",
	}
	app.connectToPubSub()

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

	wait := make(chan bool)
	<-wait
}

func NewRPCServer() *RPCServer {
	return &RPCServer{
		portReserved: make(map[int]int),
		redisClient: redis.NewClient(&redis.Options{
			Addr:     utils.URL.RedisServiceUrl,
			Password: "",
			DB:       0,
		}),
	}
}

func (app *App) connectToPubSub() error {

	var reply string
	payload := ConnectPayload{
		ClientIp: app.MyIP,
	}

	client, err := rpc.DialHTTP("tcp", utils.URL.PubSubServiceUrl)
	if err != nil {
		return err
	}

	err = client.Call("PubSub.Connect", payload, &reply)
	if err != nil {
		return err
	}

	log.Println(reply)
	return nil
}

func (app *App) listenRPC() error {
	listen, err := net.Listen("tcp", ":3000")
	if err != nil {
		return nil
	}
	defer listen.Close()

	http.Serve(listen, nil)

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(conn)
	}
}
