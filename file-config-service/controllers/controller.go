package controller

import (
	"file-config-service/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAll(c *gin.Context) {

	regions, err := models.DB.RegionEntry.GetAll()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"regions": regions,
	})
}
