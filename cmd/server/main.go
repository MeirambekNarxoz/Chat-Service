package main

import (
	"chat-service/internal/database"
	httpDelivery "chat-service/internal/delivery/http"
	wsDelivery "chat-service/internal/delivery/ws"
	"chat-service/internal/repository"
	"chat-service/internal/routes"
	"chat-service/internal/services"
	"chat-service/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	dbUrl := os.Getenv("DB_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	// Init DB and Migration
	db := database.InitDB(dbUrl)
	database.RunMigrations(dbUrl)

	// MinIO configuration
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	var minioClient *storage.MinioClient
	if minioEndpoint != "" && minioAccessKey != "" {
		minioClient = storage.NewMinioClient(minioEndpoint, minioAccessKey, minioSecretKey, minioUseSSL)
	} else {
		log.Println("WARNING: MinIO credentials missing, file upload will fail")
	}

	// Build dependencies
	chatRepo := repository.NewChatRepository(db)
	chatService := services.NewChatService(chatRepo)
	hub := services.NewHub(chatRepo)

	// Run Hub in background
	go hub.Run()

	// Handlers
	chatHandler := httpDelivery.NewChatHandler(chatService, minioClient)
	wsHandler := wsDelivery.NewWSHandler(hub, jwtSecret)

	// Setup Router
	r := gin.Default()

	routes.SetupRouter(r, chatHandler, wsHandler, jwtSecret)

	log.Printf("Chat Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
