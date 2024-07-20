package api

import (
	"broker-service/utils"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAll(c *gin.Context) {

	url := utils.URL.FileConfigServiceUrl + "/regions"
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
