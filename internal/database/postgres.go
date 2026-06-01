package database

import (
	"chat-service/internal/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(dsn string) *gorm.DB {
	if dsn == "" {
		log.Fatal("DB_URL is empty: set DB_URL or CHAT_DB_* variables in .env")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Chat{}, &models.ChatParticipant{}, &models.Message{}); err != nil {
		log.Printf("WARNING: GORM AutoMigrate failed: %v", err)
	}

	return db
}
