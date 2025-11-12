package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Println(" .env not found")
	}
	dburl := os.Getenv("DB_SERVER")

	if dburl == "" {
		log.Fatal("url database not found")
	}

	database, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		log.Fatal("err to connect", err)
	}

	fmt.Println("connected database")

	DB = database
}
