package services

import (
	"chat-service/internal/models"
	"chat-service/internal/repository"
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
			// 1. Global Chat Logic (if ChatID is 0)
			if msg.ChatID == 0 {
				msg.CreatedAt = time.Now()
				h.mu.Lock()
				msgBytes, _ := json.Marshal(msg)
				log.Printf("Global broadcast from user %d", msg.SenderID)
				for userID, conn := range h.clients {
					err := conn.WriteMessage(websocket.TextMessage, msgBytes)
					if err != nil {
						log.Printf("Error sending global msg to user %d: %v", userID, err)
						conn.Close()
						delete(h.clients, userID)
					}
				}
				h.mu.Unlock()
				continue
			}

			// 2. Private/Group Chat Logic (if ChatID > 0)
			msg.CreatedAt = time.Now()
			if err := h.chatRepo.SaveMessage(msg); err != nil {
				log.Printf("Failed to save msg to DB: %v", err)
				continue
			}

			participants, err := h.chatRepo.GetChatParticipants(msg.ChatID)
			if err != nil {
				log.Printf("Failed to get chat participants: %v", err)
				continue
			}

			h.mu.Lock()
			msgBytes, _ := json.Marshal(msg)
			for _, userID := range participants {
				if conn, ok := h.clients[userID]; ok {
					err := conn.WriteMessage(websocket.TextMessage, msgBytes)
					if err != nil {
						log.Printf("Error sending message to user %d: %v", userID, err)
						conn.Close()
						delete(h.clients, userID)
					}
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
