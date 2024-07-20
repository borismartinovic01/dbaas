package main

import (
	"encoding/json"
	"file-service/rabbit"
	"file-service/utils"
	"log"
	"net/http"
	"net/rpc"
	"os"

	"github.com/gin-gonic/gin"
)

type PublishPayload struct {
	Topic   string
	Message string
}

type NewTopicPayload struct {
	Topic string
}

type NewRegionDto struct {
	Region string `json:"region"`
}

type NewTypeDto struct {
	Type string `json:"type"`
}

type NewVersionDto struct {
	Version string `json:"version"`
}

func UploadFile(c *gin.Context) {
	region := c.Param("region")
	dbType := c.Param("type")
	version := c.Param("version")
	filename := "main.tf"
	filePath := region + "/" + dbType + "/" + version + "/" + filename

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't read file from request",
		})
		return
	}

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var reply string
	payload := PublishPayload{
		Topic:   region,
		Message: filePath,
	}

	client, err := rpc.DialHTTP("tcp", utils.URL.PubSubServiceUrl)
	if err != nil {
		log.Println("Error dialing to pubsub")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't publish on PubSub",
		})
		return
	}

	err = client.Call("PubSub.Publish", payload, &reply)
	if err != nil {
		log.Println("Error calling Publish on PubSub")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't publish on PubSub",
		})
		return
	}
	log.Println(reply)

	c.JSON(http.StatusCreated, gin.H{})
}

func GetFile(c *gin.Context) {
	region := c.Param("region")
	dbType := c.Param("type")
	version := c.Param("version")
	filename := "main.tf"
	filePath := region + "/" + dbType + "/" + version + "/" + filename

	c.File(filePath)
}

func NewRegion(c *gin.Context) {

	var requestPayload NewRegionDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	newRegion := requestPayload.Region

	err := os.MkdirAll(newRegion, 0750)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Can't create region dir",
		})
		return
	}

	var reply string
	PubSubPayload := NewTopicPayload{
		Topic: newRegion,
	}

	client, err := rpc.DialHTTP("tcp", utils.URL.PubSubServiceUrl)
	if err != nil {
		log.Println("Error dialing to pubsub")
	}

	err = client.Call("PubSub.NewTopic", PubSubPayload, &reply)
	if err != nil {
		log.Println("Error calling NewTopic on PubSub")
	}
	log.Println(reply)

	RabbitPayload := rabbit.Job{
		JobType: "REGION",
		Region:  newRegion,
	}
	body, _ := json.Marshal(RabbitPayload)

	err = Publisher.Push(body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func NewDatabaseType(c *gin.Context) {
	region := c.Param("region")

	var requestPayload NewTypeDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	dbType := requestPayload.Type
	directoryPath := region + "/" + dbType

	err := os.Mkdir(directoryPath, 0750)
	if err != nil && !os.IsExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Can't create type dir",
		})
		return
	}

	RabbitPayload := rabbit.Job{
		JobType: "TYPE",
		Region:  region,
		Type:    dbType,
	}
	body, _ := json.Marshal(RabbitPayload)

	err = Publisher.Push(body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func NewVersion(c *gin.Context) {
	region := c.Param("region")
	dbType := c.Param("type")

	var requestPayload NewVersionDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	version := requestPayload.Version
	directoryPath := region + "/" + dbType + "/" + version

	err := os.Mkdir(directoryPath, 0750)
	if err != nil && !os.IsExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Can't create version dir",
		})
		return
	}

	RabbitPayload := rabbit.Job{
		JobType: "VERSION",
		Region:  region,
		Type:    dbType,
		Version: version,
	}
	body, _ := json.Marshal(RabbitPayload)

	err = Publisher.Push(body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}
