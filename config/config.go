package config

import (
	"fmt"
	"go-billing-engine/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
		return err
	}
	log.Println(".env file loaded successfully")
	return nil
}

func ConnectDatabase() {
	if err := LoadEnv(); err != nil {
		log.Println("Continuing without .env file")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASS", ""),
		getEnv("DB_NAME", "billing"),
		getEnv("DB_PORT", "5432"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Pricing{},
		&models.Loan{},
		&models.Installment{},
		&models.Payment{},
	); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	DB = db
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
