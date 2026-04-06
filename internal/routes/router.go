package routes

import (
	httpDelivery "chat-service/internal/delivery/http"
	wsDelivery "chat-service/internal/delivery/ws"
	"chat-service/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	r *gin.Engine,
	chatHandler *httpDelivery.ChatHandler,
	wsHandler *wsDelivery.WSHandler,
	jwtSecret string,
) {

	// Public / Websocket upgrades
	r.GET("/api/chat/ws", wsHandler.Connect)

	// Protected REST API
	apiParams := r.Group("/api/chat")
	apiParams.Use(middleware.AuthMiddleware(jwtSecret))
	{
		apiParams.POST("/create", chatHandler.CreateChat)
		apiParams.GET("/history/:chat_id", chatHandler.GetHistory)
		apiParams.POST("/messages/upload", chatHandler.UploadFile)
		apiParams.GET("/list", chatHandler.GetUserChats)
	}
}
