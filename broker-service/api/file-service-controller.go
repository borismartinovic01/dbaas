package api

import (
	"broker-service/dto"
	"broker-service/utils"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadFile(c *gin.Context) {
	region := c.Param("region")
	dbType := c.Param("type")
	version := c.Param("version")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Can't read file from request",
		})
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err = prepareFileForRequest(file, writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to handle uploaded file",
		})
		return
	}
	writer.Close()

	url := utils.URL.FileServiceUrl + "/regions/" + region + "/types/" + dbType + "/versions/" + version
	request, err := http.NewRequest("POST", url, body)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	c.JSON(response.StatusCode, response.Body)
}

func NewRegion(c *gin.Context) {

	var requestPayload dto.NewRegionDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(requestPayload); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := utils.URL.FileServiceUrl + "/regions"
	request, err := http.NewRequest("POST", url, &buffer)
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

	c.JSON(response.StatusCode, response.Body)
}

func NewDatabaseType(c *gin.Context) {
	region := c.Param("region")
	var requestPayload dto.NewTypeDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(requestPayload); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := utils.URL.FileServiceUrl + "/regions/" + region + "/types"
	request, err := http.NewRequest("POST", url, &buffer)
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

	c.JSON(response.StatusCode, response.Body)
}

func NewVersion(c *gin.Context) {
	region := c.Param("region")
	dbType := c.Param("type")
	var requestPayload dto.NewVersionDto

	if err := c.BindJSON(&requestPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read the body"})
		return
	}

	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)

	if err := encoder.Encode(requestPayload); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	url := utils.URL.FileServiceUrl + "/regions/" + region + "/types/" + dbType + "/versions"
	request, err := http.NewRequest("POST", url, &buffer)
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

	c.JSON(response.StatusCode, response.Body)
}

func prepareFileForRequest(file *multipart.FileHeader, writer *multipart.Writer) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(part, src)
	if err != nil {
		return err
	}

	return nil
}
