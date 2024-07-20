package main

import (
	"broker-service/api"
	"broker-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	utils.InitUrl()
	utils.CreateRedisClient()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Next()
	})

	router.POST("/signup", api.Signup)
	router.POST("/login", api.Login)

	router.POST("/servers", api.AuthenticateUser, api.CreateServer)
	router.POST("/databases", api.AuthenticateUser, api.CreateDatabase)
	router.GET("/users/databases/:name", api.AuthenticateUser, api.DbOverviewByName)
	router.GET("/users/databases", api.AuthenticateUser, api.UserDatabases)
	router.GET("/users/servers", api.AuthenticateUser, api.UserServers)
	router.GET("/users/databases/:name/grafana", api.AuthenticateUser, api.DbGrafanaUIDByName)
	router.GET("/deployments/:uuid", api.DeploymentStatus)

	router.POST("/regions/:region/types/:type/versions/:version", api.AuthenticateAdmin, api.UploadFile)
	router.POST("/regions", api.AuthenticateAdmin, api.NewRegion)
	router.POST("/regions/:region/types", api.AuthenticateAdmin, api.NewDatabaseType)
	router.POST("/regions/:region/types/:type/versions", api.AuthenticateAdmin, api.NewVersion)

	router.GET("/regions", api.Authenticate, api.GetAll)

	router.POST("/subscriptions", api.AuthenticateAdmin, api.Subscribe)

	router.Run()
}
