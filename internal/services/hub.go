package services

import (
	"chat-service/internal/models"
	"chat-service/internal/repository"
	"encoding/json"
	"log"
	"sync"
	"time"
)

type Hub struct {
	clients    map[uint]chan []byte
	register   chan *ClientInfo
	unregister chan uint
	broadcast  chan *models.Message
	mu         sync.Mutex

	chatRepo *repository.ChatRepository
}

type ClientInfo struct {
	UserID uint
	Send   chan []byte
}

func NewHub(chatRepo *repository.ChatRepository) *Hub {
	return &Hub{
		clients:    make(map[uint]chan []byte),
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
			if oldCh, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(oldCh) // Disconnect previous session
			}
			h.clients[client.UserID] = client.Send
			h.mu.Unlock()
			log.Printf("User %d connected", client.UserID)

			// Broadcast presence update (outside lock)
			h.broadcastPresence(client.UserID, true)

		case userID := <-h.unregister:
			h.mu.Lock()
			if sendCh, ok := h.clients[userID]; ok {
				close(sendCh)
				delete(h.clients, userID)
				log.Printf("User %d disconnected", userID)
				h.mu.Unlock() // Unlock before broadcasting
				h.broadcastPresence(userID, false)
			} else {
				h.mu.Unlock()
			}

		case msg := <-h.broadcast:
			// 1. Global Chat Logic (if ChatID is 0)
			if msg.ChatID == 0 {
				msg.CreatedAt = time.Now()
				h.mu.Lock()
				msgBytes, _ := json.Marshal(msg)
				log.Printf("Global broadcast from user %d", msg.SenderID)
				for userID, sendCh := range h.clients {
					select {
					case sendCh <- msgBytes:
					default:
						// Buffer full - assumes client is dead or too slow
						close(sendCh)
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
				if sendCh, ok := h.clients[userID]; ok {
					select {
					case sendCh <- msgBytes:
					default:
						close(sendCh)
						delete(h.clients, userID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) RegisterClient(userID uint, sendCh chan []byte) {
	h.register <- &ClientInfo{UserID: userID, Send: sendCh}
}

func (h *Hub) UnregisterClient(userID uint) {
	h.unregister <- userID
}

func (h *Hub) BroadcastMessage(msg *models.Message) {
	h.broadcast <- msg
}

func (h *Hub) broadcastPresence(senderID uint, online bool) {
	status := "PRESENCE_OFFLINE"
	if online {
		status = "PRESENCE_ONLINE"
	}

	msg := &models.Message{
		ChatID:    0,
		SenderID:  senderID,
		Type:      status,
		CreatedAt: time.Now(),
	}
	msgBytes, _ := json.Marshal(msg)

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, sendCh := range h.clients {

		select {
		case sendCh <- msgBytes:
		default:
			// Buffer full - assumes client is dead or too slow
			// (Don't delete here to avoid modifying map during iteration)
		}
	}
}

func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	_, online := h.clients[userID]
	return online
}
