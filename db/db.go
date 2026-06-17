package db

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = sql.Open("mysql", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Auth Service connected to MySQL")
}
