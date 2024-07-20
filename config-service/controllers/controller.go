package controllers

import (
	"config-service/dto"
	"config-service/models"
	"config-service/node-info"
	"log"
	"net/http"
	"net/rpc"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CreateDatabasePayload struct {
	Name     string
	Type     string
	Version  string
	Password string
	User     string
	UUID     uuid.UUID
}

type CreateDatabaseResponse struct {
	Status   string
	NodeIP   string
	NodePort string
}

func CreateServer(c *gin.Context) {
	var serverDto dto.ServerDto

	if err := c.BindJSON(&serverDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(serverDto.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash the password",
		})
		return
	}

	serverEntry := models.ServerEntry{
		Name:      serverDto.Name,
		Location:  serverDto.Location,
		Admin:     serverDto.Admin,
		Password:  string(hash),
		Email:     serverDto.Email,
		CreatedAt: time.Now(),
		Status:    "UP",
	}

	err = models.DB.ServerEntry.Insert(serverEntry)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusCreated, gin.H{})
}

func CreateDatabase(c *gin.Context) {
	var databaseDto dto.DatabaseDto

	if err := c.BindJSON(&databaseDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	server, err := models.DB.ServerEntry.GetOne(databaseDto.Server)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	directoryUUID := uuid.New()
	go createDatabase(databaseDto, server, directoryUUID)

	c.JSON(http.StatusCreated, gin.H{
		"uuid": directoryUUID.String(),
	})
}

func createDatabase(databaseDto dto.DatabaseDto, server *models.ServerEntry, directoryUUID uuid.UUID) {

	var reply CreateDatabaseResponse
	payload := CreateDatabasePayload{
		Name:     databaseDto.Name,
		Type:     databaseDto.Type,
		Version:  databaseDto.Version,
		User:     server.Admin,
		Password: databaseDto.Password,
		UUID:     directoryUUID,
	}

	client, err := rpc.DialHTTP("tcp", node.LocationServerMp[server.Location])
	if err != nil {
		log.Println("Error when dialing node rpc")
		return
	}

	err = client.Call("RPCServer.CreateDatabase", payload, &reply)
	if err != nil {
		log.Println("Error when calling node rpc")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(databaseDto.Password), 10)
	if err != nil {
		log.Println("Error: Failed to hash the password")
		return
	}

	database := models.DatabaseEntry{
		Name:        databaseDto.Name,
		Password:    string(hash),
		Server:      databaseDto.Server,
		Environment: databaseDto.Environment,
		Configuration: models.Configuration{
			ServiceType:     databaseDto.ServiceType,
			ComputeType:     databaseDto.ComputeType,
			MaxStorageSize:  databaseDto.MaxStorageSize,
			StorageSizeUnit: databaseDto.StorageSizeUnit,
		},
		Connectivity:  databaseDto.Connectivity,
		Type:          databaseDto.Type,
		Version:       databaseDto.Version,
		NodeIP:        strings.Split(reply.NodeIP, ":")[0],
		NodePort:      reply.NodePort,
		DirectoryUUID: directoryUUID.String(),
		GrafanaUID:    "",
		Email:         databaseDto.Email,
		CreatedAt:     time.Now(),
		Status:        "ONLINE",
	}

	err = models.DB.DatabaseEntry.Insert(database)
	if err != nil {
		log.Println("Error: Failed to insert new database entry")
		return
	}
}

func DbOverviewByName(c *gin.Context) {
	email := c.Param("email")
	name := c.Param("name")

	database, err := models.DB.DatabaseEntry.GetOne(name, email)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	server, err := models.DB.ServerEntry.GetOne(database.Server)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response := dto.DatabaseOverviewDto{
		Status:        database.Status,
		Location:      server.Location,
		Server:        database.Server,
		Environment:   database.Environment,
		Connectivity:  database.Connectivity,
		Type:          database.Type,
		Version:       database.Version,
		Configuration: dto.ConfigurationDto(database.Configuration),
		NodeIP:        database.NodeIP,
		NodePort:      database.NodePort,
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func UserDatabases(c *gin.Context) {
	email := c.Param("email")

	databases, err := models.DB.DatabaseEntry.GetAllByEmail(email)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var response []dto.ResourceDto

	for _, v := range databases {

		server, err := models.DB.ServerEntry.GetOne(v.Server)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		resource := dto.ResourceDto{
			Name:     v.Name,
			Location: server.Location,
			Server:   v.Server,
		}

		response = append(response, resource)
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func UserServers(c *gin.Context) {
	email := c.Param("email")

	servers, err := models.DB.ServerEntry.GetAll(email)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var response []dto.ServerResponseDto

	for _, v := range servers {

		resource := dto.ServerResponseDto{
			Name:     v.Name,
			Location: v.Location,
		}

		response = append(response, resource)
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func DbGrafanaUIDByName(c *gin.Context) {
	email := c.Param("email")
	name := c.Param("name")

	database, err := models.DB.DatabaseEntry.GetOne(name, email)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	response := dto.DatabaseGrafanaDto{
		GrafanaUID:    database.GrafanaUID,
		DirectoryUUID: database.DirectoryUUID,
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
	})
}

func UpdateGrafanaUIDByDirectoryUUID(c *gin.Context) {

	var setGrafanaDto dto.SetGrafanaDto

	if err := c.BindJSON(&setGrafanaDto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	directoryUUID := c.Param("directoryUUID")

	err := models.DB.DatabaseEntry.UpdateGrafanaUID(directoryUUID, setGrafanaDto.GrafanaUID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
