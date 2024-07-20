package api

import (
	"broker-service/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Signup(c *gin.Context) {
	jsonData := c.Request.Body

	signupUrl := utils.URL.AuthenticationServiceUrl + "/signup"
	request, err := http.NewRequest("POST", signupUrl, jsonData)
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

func Login(c *gin.Context) {
	jsonData := c.Request.Body

	loginUrl := utils.URL.AuthenticationServiceUrl + "/login"
	request, err := http.NewRequest("POST", loginUrl, jsonData)
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

	var tokenResponse struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(response.Body).Decode(&tokenResponse)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	c.JSON(response.StatusCode, gin.H{
		"token": tokenResponse.Token,
	})
}

func AuthenticateUser(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) || claims["role"].(string) != "user" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func AuthenticateAdmin(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) || claims["role"].(string) != "admin" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}

func Authenticate(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
	})
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
