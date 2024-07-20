package api

import (
	"broker-service/utils"
	"log"
	"net/http"
	"net/rpc"

	"github.com/gin-gonic/gin"
)

type SubscribePayload struct {
	Topic    string
	ClientIp string
}

type SubscribeDto struct {
	Topic    string `json:"topic"`
	ClientIp string `json:"clientIp"`
}

func Subscribe(c *gin.Context) {
	var requestPayload SubscribeDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Can't bind request body"})
		return
	}

	var reply string
	payload := SubscribePayload{
		Topic:    requestPayload.Topic,
		ClientIp: requestPayload.ClientIp,
	}

	client, err := rpc.DialHTTP("tcp", utils.URL.PubSubServiceUrl)
	if err != nil {
		log.Println("Error dialing to pubsub")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't subscribe on PubSub",
		})
		return
	}

	err = client.Call("PubSub.Subscribe", payload, &reply)
	if err != nil {
		log.Println("Error calling Subscribe on PubSub")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't subscribe on PubSub",
		})
		return
	}
	log.Println(reply)

	c.JSON(http.StatusOK, gin.H{})
}
