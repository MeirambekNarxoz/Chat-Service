package main

import (
	"chat-service/internal/config"
	"chat-service/internal/database"
	httpDelivery "chat-service/internal/delivery/http"
	wsDelivery "chat-service/internal/delivery/ws"
	"chat-service/internal/repository"
	"chat-service/internal/routes"
	"chat-service/internal/services"
	"chat-service/internal/storage"

	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	// Init DB and Run Migrations
	database.RunMigrations(cfg.DBURL)
	db := database.InitDB(cfg.DBURL)

	var minioClient *storage.MinioClient
	if cfg.MinIOEndpoint != "" && cfg.MinIOAccessKey != "" {
		minioClient = storage.NewMinioClient(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOUseSSL)
	} else {
		log.Println("WARNING: MinIO credentials missing, file upload will fail")
	}

	// Build dependencies
	chatRepo := repository.NewChatRepository(db)
	hub := services.NewHub(chatRepo)
	chatService := services.NewChatService(chatRepo, hub)

	// Run Hub in background
	go hub.Run()

	// Handlers
	chatHandler := httpDelivery.NewChatHandler(chatService, minioClient)
	wsHandler := wsDelivery.NewWSHandler(hub, cfg.JwtSecret)

	// Setup Router
	r := gin.Default()
	routes.SetupRouter(r, chatHandler, wsHandler, cfg.JwtSecret)

	log.Printf("Chat Service starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
