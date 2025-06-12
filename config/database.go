package config

import (
	"dorm-chat-api/models" // не забудь про импорт своих моделей!
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=require", // БЫЛО: disable, СТАЛО: require
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}

// Миграция только для Go-моделей
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&models.Chat{}, &models.Message{})
}
