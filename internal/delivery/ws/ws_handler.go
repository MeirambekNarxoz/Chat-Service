package ws

import (
	"chat-service/internal/middleware"
	"chat-service/internal/models"
	"chat-service/internal/services"
	"encoding/json"
	stdhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub       *services.Hub
	jwtSecret string
	upgrader  websocket.Upgrader
}

func NewWSHandler(hub *services.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *stdhttp.Request) bool { return true },
		},
	}
}

func (h *WSHandler) Connect(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(stdhttp.StatusUnauthorized, gin.H{"error": "token is required"})
		return
	}

	userID, err := middleware.UserIDFromToken(token, h.jwtSecret)
	if err != nil {
		c.JSON(stdhttp.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// upgrader handles the error response
		return
	}

	h.hub.RegisterClient(userID, conn)

	// Clean up on disconnect
	defer func() {
		h.hub.UnregisterClient(userID)
		conn.Close()
	}()

	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg models.Message
		if err := json.Unmarshal(messageData, &msg); err == nil {
			msg.SenderID = userID
			h.hub.BroadcastMessage(&msg)
		}
	}
}
