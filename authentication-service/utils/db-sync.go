package utils

import (
	"authentication-service/models"

	"golang.org/x/crypto/bcrypt"
)

func SyncDB() {
	hash, _ := bcrypt.GenerateFromPassword([]byte("martinovicboris01"), 10)
	admin := models.User{
		Email:    "martinovicboris01@gmail.com",
		Password: string(hash),
		Role:     "admin",
	}

	var checkAdmin models.User
	resultCheck := DB.Where("email = ?", admin.Email).First(&checkAdmin)
	if resultCheck.Error != nil {
		DB.Create(&admin)
	}
}
