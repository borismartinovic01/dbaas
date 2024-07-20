package utils

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	dsn := os.Getenv("DSN")
	connCounts := 0
	maxCounts := 20

	for {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

		if err != nil {
			log.Println("Postgres not ready yet...")
			connCounts++
		} else {
			log.Println("Connected to postgres.")
			return
		}

		if connCounts > maxCounts {
			panic("Can't connect to postgres")
		}

		time.Sleep(2 * time.Second)
		continue
	}
}
