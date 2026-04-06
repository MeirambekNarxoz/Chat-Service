package services

import (
	"chat-service/internal/models"
	"chat-service/internal/repository"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[uint]*websocket.Conn
	register   chan *ClientInfo
	unregister chan uint
	broadcast  chan *models.Message
	mu         sync.Mutex

	chatRepo *repository.ChatRepository
}

type ClientInfo struct {
	UserID uint
	Conn   *websocket.Conn
}

func NewHub(chatRepo *repository.ChatRepository) *Hub {
	return &Hub{
		clients:    make(map[uint]*websocket.Conn),
		register:   make(chan *ClientInfo),
		unregister: make(chan uint),
		broadcast:  make(chan *models.Message),
		chatRepo:   chatRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client.Conn
			h.mu.Unlock()
			log.Printf("User %d connected", client.UserID)

		case userID := <-h.unregister:
			h.mu.Lock()
			if conn, ok := h.clients[userID]; ok {
				conn.Close()
				delete(h.clients, userID)
				log.Printf("User %d disconnected", userID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			// Save to DB
			msg.CreatedAt = time.Now()
			if err := h.chatRepo.SaveMessage(msg); err != nil {
				log.Printf("Failed to save msg: %v", err)
				continue
			}

			// Ideally, you get all participants of the chat, and only send to them
			// For simplicity we fetch online users. In prod, you use Redis Pub/Sub here.
			// Let's pretend everyone is in the chat
			h.mu.Lock()
			for _, conn := range h.clients {
				msgBytes, _ := json.Marshal(msg)
				err := conn.WriteMessage(websocket.TextMessage, msgBytes)
				if err != nil {
					log.Printf("Error sending message: %v", err)
					conn.Close()
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) RegisterClient(userID uint, conn *websocket.Conn) {
	h.register <- &ClientInfo{UserID: userID, Conn: conn}
}

func (h *Hub) UnregisterClient(userID uint) {
	h.unregister <- userID
}

func (h *Hub) BroadcastMessage(msg *models.Message) {
	h.broadcast <- msg
}
