package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kalpesh-kashyap/video-streaming/services/video-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDb() {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error in loading .env file", err)
	}

	dbHost := os.Getenv("host")
	dbPort := os.Getenv("port")
	dbUser := os.Getenv("user")
	dbPassword := os.Getenv("password")
	dbName := os.Getenv("dbname")

	dns := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	DB, err = gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}
	fmt.Println("Database connection established!")

}

func MiggrateDb() {
	err := DB.AutoMigrate(&models.Video{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	fmt.Println("Database migration completed!")
}
