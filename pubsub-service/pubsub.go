package main

import (
	"log"
	"net/rpc"
	"sync"
)

type PubSub struct {
	connections               []string
	topicSubscribers          map[string][]string
	topicMtx                  sync.Mutex
	topicMessage              map[string]string
	topicSubscribersMtx       map[string]*sync.Mutex
	topicMessageMtx           map[string]*sync.Mutex
	connectionsMtx            sync.Mutex
	subscribersDeadLetters    map[string][]DeadLetterPair
	deadLettersMtx            sync.Mutex
	subscribersDeadLettersMtx map[string]*sync.Mutex
}

type DeadLetterPair struct {
	Topic   string
	Message string
}

func NewPubSub() *PubSub {
	return &PubSub{
		connections:               make([]string, 0),
		topicSubscribers:          make(map[string][]string),
		topicMtx:                  sync.Mutex{},
		topicMessage:              make(map[string]string),
		topicSubscribersMtx:       make(map[string]*sync.Mutex),
		topicMessageMtx:           make(map[string]*sync.Mutex),
		connectionsMtx:            sync.Mutex{},
		subscribersDeadLetters:    make(map[string][]DeadLetterPair),
		deadLettersMtx:            sync.Mutex{},
		subscribersDeadLettersMtx: make(map[string]*sync.Mutex),
	}
}

type ConnectPayload struct {
	ClientIp string
}

type SubscribePayload struct {
	Topic    string
	ClientIp string
}

type NewTopicPayload struct {
	Topic string
}

type EmptyPayload struct{}

type PublishPayload struct {
	Topic   string
	Message string
}

type SendDeadLettersPayload struct {
	DeadLetters []DeadLetterPair
}

type SendMessagePayload struct {
	Message string
}

func (pubsub *PubSub) Connect(payload ConnectPayload, reply *string) error {

	clientIp := payload.ClientIp

	pubsub.connectionsMtx.Lock()

	for _, conn := range pubsub.connections {
		if conn == clientIp {
			*reply = "CONNECTED"
			log.Println("Client: " + clientIp + " already connected.")
			pubsub.connectionsMtx.Unlock()
			return nil
		}
	}
	pubsub.connections = append(pubsub.connections, clientIp)
	pubsub.connectionsMtx.Unlock()

	pubsub.deadLettersMtx.Lock()
	if _, exists := pubsub.subscribersDeadLetters[clientIp]; !exists {
		pubsub.subscribersDeadLetters[clientIp] = []DeadLetterPair{}
		pubsub.subscribersDeadLettersMtx[clientIp] = &sync.Mutex{}
	}
	pubsub.deadLettersMtx.Unlock()

	dlPairs := pubsub.getDeadLetters(clientIp)
	if len(dlPairs) > 0 {
		pubsub.sendDeadLetters(clientIp, dlPairs)
	}

	log.Println("Client: " + clientIp + " successfully connected.")
	*reply = "CONNECTED"
	return nil
}

func (pubsub *PubSub) sendDeadLetters(clientIp string, dlPairs []DeadLetterPair) {
	var reply string
	payload := SendDeadLettersPayload{
		DeadLetters: dlPairs,
	}

	client, err := rpc.DialHTTP("tcp", clientIp)
	if err != nil {
		log.Println(err)
		return
	}

	err = client.Call("RPCServer.SendDeadLetters", payload, &reply)
	if err != nil {
		log.Println(err)
	}
}

func (pubsub *PubSub) Subscribe(payload SubscribePayload, reply *string) error {

	topic := payload.Topic
	clientIp := payload.ClientIp
	connExists := false

	pubsub.connectionsMtx.Lock()
	for _, conn := range pubsub.connections {
		if conn == clientIp {
			connExists = true
			break
		}
	}

	if !connExists {
		*reply = "ERR: NOT CONNECTED"
		pubsub.connectionsMtx.Unlock()
		return nil
	}
	pubsub.connectionsMtx.Unlock()

	pubsub.topicMtx.Lock()
	if _, exists := pubsub.topicSubscribers[topic]; !exists {
		pubsub.topicMtx.Unlock()
		*reply = "ERR: NO TOPIC"
		return nil
	}
	pubsub.topicMtx.Unlock()

	pubsub.topicSubscribersMtx[topic].Lock()
	defer pubsub.topicSubscribersMtx[topic].Unlock()
	for _, currIp := range pubsub.topicSubscribers[topic] {
		if currIp == clientIp {
			*reply = "SUBSCRIBED"
			return nil
		}
	}

	pubsub.topicSubscribers[topic] = append(pubsub.topicSubscribers[topic], clientIp)

	*reply = "SUBSCRIBED"
	return nil
}

func (pubsub *PubSub) Unsubscribe(payload SubscribePayload, reply *string) error {
	pubsub.removeSubscriberFromTopic(payload.Topic, payload.ClientIp)
	pubsub.removeMessagesFromQueue(payload.Topic, payload.ClientIp)

	*reply = "UNSUBSCRIBED"
	return nil
}

func (pubsub *PubSub) removeSubscriberFromTopic(topic string, clientIp string) {
	pubsub.topicSubscribersMtx[topic].Lock()
	defer pubsub.topicSubscribersMtx[topic].Unlock()

	var newTopicSubscribers []string

	for _, subscriber := range pubsub.topicSubscribers[topic] {
		if subscriber != clientIp {
			newTopicSubscribers = append(newTopicSubscribers, subscriber)
		}
	}

	pubsub.topicSubscribers[topic] = newTopicSubscribers
}

func (pubsub *PubSub) removeMessagesFromQueue(topic, clientIp string) {
	pubsub.subscribersDeadLettersMtx[clientIp].Lock()
	defer pubsub.subscribersDeadLettersMtx[clientIp].Unlock()

	var newSubscribersDeadLetters []DeadLetterPair

	for _, dlPair := range pubsub.subscribersDeadLetters[clientIp] {
		if dlPair.Topic != topic {
			newSubscribersDeadLetters = append(newSubscribersDeadLetters, dlPair)
		}
	}

	pubsub.subscribersDeadLetters[clientIp] = newSubscribersDeadLetters
}

func (pubsub *PubSub) NewTopic(payload NewTopicPayload, reply *string) error {

	topic := payload.Topic

	pubsub.topicMtx.Lock()
	defer pubsub.topicMtx.Unlock()

	if _, exists := pubsub.topicSubscribers[topic]; exists {
		log.Println("Topic " + topic + " already exists.")
		*reply = "CREATED"
		return nil
	}

	pubsub.topicSubscribers[topic] = []string{}
	pubsub.topicSubscribersMtx[topic] = &sync.Mutex{}
	pubsub.topicMessage[topic] = ""
	pubsub.topicMessageMtx[topic] = &sync.Mutex{}

	log.Println("Topic " + topic + " created.")
	*reply = "CREATED"
	return nil
}

func (pubsub *PubSub) Publish(payload PublishPayload, reply *string) error {

	topic := payload.Topic
	message := payload.Message

	pubsub.topicMessageMtx[topic].Lock()
	defer pubsub.topicMessageMtx[topic].Unlock()

	pubsub.topicMessage[topic] = message

	var wg sync.WaitGroup

	pubsub.topicSubscribersMtx[topic].Lock()

	wg.Add(len(pubsub.topicSubscribers[topic]))
	for _, subscriber := range pubsub.topicSubscribers[topic] {
		go pubsub.sendMessage(topic, subscriber, &wg)
	}

	pubsub.topicSubscribersMtx[topic].Unlock()

	wg.Wait()

	*reply = "PUBLISHED"
	return nil
}

func (pubsub *PubSub) sendMessage(topic string, subscriber string, wg *sync.WaitGroup) {
	message := pubsub.topicMessage[topic]
	wg.Done()

	pubsub.connectionsMtx.Lock()
	connectionExists := false

	for _, conn := range pubsub.connections {
		if conn == subscriber {
			connectionExists = true
			break
		}
	}
	pubsub.connectionsMtx.Unlock()

	if connectionExists {

		var reply string
		payload := SendMessagePayload{
			Message: message,
		}

		client, err := rpc.DialHTTP("tcp", subscriber)
		if err != nil {
			log.Println(err)
			pubsub.addDeadLetter(topic, subscriber, message)
			pubsub.removeConnection(subscriber)
			return
		}

		err = client.Call("RPCServer.SendMessage", payload, &reply)
		if err != nil {
			log.Println(err)
			pubsub.addDeadLetter(topic, subscriber, message)
			pubsub.removeConnection(subscriber)
			return
		}

		return
	}

	pubsub.addDeadLetter(topic, subscriber, message)
}

func (pubsub *PubSub) removeConnection(subscriber string) {
	pubsub.connectionsMtx.Lock()
	defer pubsub.connectionsMtx.Unlock()

	newConnections := make([]string, 0)

	for _, conn := range pubsub.connections {
		if conn != subscriber {
			newConnections = append(newConnections, conn)
		}
	}

	pubsub.connections = newConnections
}

func (pubsub *PubSub) addDeadLetter(topic, subscriber, message string) {
	pubsub.subscribersDeadLettersMtx[subscriber].Lock()

	pubsub.subscribersDeadLetters[subscriber] = append(
		pubsub.subscribersDeadLetters[subscriber],
		DeadLetterPair{Topic: topic, Message: message},
	)

	pubsub.subscribersDeadLettersMtx[subscriber].Unlock()
}

func (pubsub *PubSub) getDeadLetters(subscriber string) []DeadLetterPair {

	pubsub.deadLettersMtx.Lock()
	if _, exists := pubsub.subscribersDeadLetters[subscriber]; !exists {
		pubsub.deadLettersMtx.Unlock()
		return []DeadLetterPair{}
	}
	pubsub.deadLettersMtx.Unlock()

	result := make([]DeadLetterPair, 0)
	topics := make(map[string]bool)

	pubsub.subscribersDeadLettersMtx[subscriber].Lock()
	defer pubsub.subscribersDeadLettersMtx[subscriber].Unlock()

	for _, dlPair := range pubsub.subscribersDeadLetters[subscriber] {
		topics[dlPair.Topic] = false
	}

	for topic := range topics {
		for _, dlPair := range pubsub.subscribersDeadLetters[subscriber] {
			if dlPair.Topic == topic {
				result = append(result, dlPair)
			}
		}
	}

	pubsub.subscribersDeadLetters[subscriber] = nil
	return result
}
