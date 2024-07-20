package api

import (
	"broker-service/dto"
	"broker-service/utils"
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateServer(c *gin.Context) {
	var requestPayload dto.ServerDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read the body"})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	userEmail := utils.GetEmailFromJwt(tokenString)

	requestPayload.Email = userEmail

	var buffer bytes.Buffer

	encoder := json.NewEncoder(&buffer)
	if err := encoder.Encode(requestPayload); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	createServerUrl := utils.URL.ConfigServiceUrl + "/servers"
	request, err := http.NewRequest("POST", createServerUrl, &buffer)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	c.JSON(response.StatusCode, gin.H{})
}

func CreateDatabase(c *gin.Context) {
	var requestPayload dto.DatabaseDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userEmail := utils.GetEmailFromJwt(tokenString)
	requestPayload.Email = userEmail

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(requestPayload); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	createDatabaseUrl := utils.URL.ConfigServiceUrl + "/databases"
	request, err := http.NewRequest("POST", createDatabaseUrl, &buffer)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", response.Header.Get("Content-Type"))
	c.Data(response.StatusCode, response.Header.Get("Content-Type"), body)
}

func DbOverviewByName(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userEmail := utils.GetEmailFromJwt(tokenString)

	url := utils.URL.ConfigServiceUrl + "/users/" + userEmail + "/databases/" + c.Param("name")
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", response.Header.Get("Content-Type"))
	c.Data(response.StatusCode, response.Header.Get("Content-Type"), body)
}

func DbGrafanaUIDByName(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userEmail := utils.GetEmailFromJwt(tokenString)

	url := utils.URL.ConfigServiceUrl + "/users/" + userEmail + "/databases/" + c.Param("name") + "/grafana"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", response.Header.Get("Content-Type"))
	c.Data(response.StatusCode, response.Header.Get("Content-Type"), body)
}

func UserDatabases(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userEmail := utils.GetEmailFromJwt(tokenString)

	url := utils.URL.ConfigServiceUrl + "/users/" + userEmail + "/databases"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", response.Header.Get("Content-Type"))
	c.Data(response.StatusCode, response.Header.Get("Content-Type"), body)
}

func UserServers(c *gin.Context) {

	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userEmail := utils.GetEmailFromJwt(tokenString)

	url := utils.URL.ConfigServiceUrl + "/users/" + userEmail + "/servers"
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", response.Header.Get("Content-Type"))
	c.Data(response.StatusCode, response.Header.Get("Content-Type"), body)
}

func DeploymentStatus(c *gin.Context) {
	deploymentUUID := c.Param("uuid")

	deploymentStatus := utils.RedisClient.Get(deploymentUUID)
	result, _ := deploymentStatus.Result()
	c.JSON(http.StatusOK, gin.H{
		"status": result,
	})
}
