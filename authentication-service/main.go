package main

import (
	"authentication-service/controllers"
	"authentication-service/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	utils.ConnectToDB()
	utils.SyncDB()

	router := gin.Default()

	router.POST("/login", controllers.Login)
	router.POST("/signup", controllers.Signup)

	router.Run()
}
