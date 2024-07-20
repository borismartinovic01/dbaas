package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "192.168.1.11"
	port     = "3005"
	user     = "admin"
	password = "pass"
	dbname   = "wer3"
)

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Unable to connect to database")
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Unable to verify connection")
	}
	fmt.Println("Successfully connected to the database!")

	createTableSQL := "CREATE TABLE IF NOT EXISTS test_table (id SERIAL PRIMARY KEY, data VARCHAR(100))"
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal("Unable to create table")
	}
	fmt.Println("Table created or already exists")

	for i := 0; i < 1000; i++ {

		insertSQL := "INSERT INTO test_table (data) VALUES ($1)"
		_, err := db.Exec(insertSQL, fmt.Sprintf("test_data_%d", i))
		if err != nil {
			log.Fatal("Unable to insert data")
		}

		selectSQL := "SELECT * FROM test_table"
		rows, err := db.Query(selectSQL)
		if err != nil {
			log.Fatal("Unable to select data:")
		}

		for rows.Next() {
			var id int
			var data string
			if err := rows.Scan(&id, &data); err != nil {
				log.Fatal("Unable to scan row")
			}
		}
		rows.Close()

		time.Sleep(100 * time.Millisecond)
	}
}
