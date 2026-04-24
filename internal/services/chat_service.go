package services

import (
	"chat-service/internal/models"
	"chat-service/internal/repository"
)

type ChatService struct {
	repo *repository.ChatRepository
	hub  *Hub
}

func NewChatService(repo *repository.ChatRepository, hub *Hub) *ChatService {
	return &ChatService{repo: repo, hub: hub}
}

type ChatDTO struct {
	ID          uint            `json:"chat_id"`
	RecipientID uint            `json:"recipient_id"`
	LastMessage *models.Message `json:"last_message"`
	UnreadCount int64           `json:"unread_count"`
	IsOnline    bool            `json:"is_online"`
}

func (s *ChatService) GetOrCreatePersonalChat(user1ID, user2ID uint) (*models.Chat, error) {
	// Сначала ищем существующий
	chat, err := s.repo.GetPersonalChat(user1ID, user2ID)
	if err == nil {
		return chat, nil
	}

	// Если не нашли — создаем
	return s.CreatePersonalChat(user1ID, user2ID)
}

func (s *ChatService) CreatePersonalChat(user1ID, user2ID uint) (*models.Chat, error) {
	chat := &models.Chat{}
	if err := s.repo.CreateChat(chat); err != nil {
		return nil, err
	}

	_ = s.repo.AddParticipant(&models.ChatParticipant{ChatID: chat.ID, UserID: user1ID})
	_ = s.repo.AddParticipant(&models.ChatParticipant{ChatID: chat.ID, UserID: user2ID})

	return chat, nil
}

func (s *ChatService) GetHistory(chatID uint) ([]models.Message, error) {
	return s.repo.GetMessagesByChatID(chatID)
}

func (s *ChatService) MarkAsRead(chatID, userID uint) error {
	return s.repo.MarkAsRead(chatID, userID)
}

func (s *ChatService) GetUserChatsRich(userID uint) ([]ChatDTO, error) {
	chats, err := s.repo.GetUserChats(userID)
	if err != nil {
		return nil, err
	}

	if len(chats) == 0 {
		return []ChatDTO{}, nil
	}

	var chatIDs []uint
	for _, c := range chats {
		chatIDs = append(chatIDs, c.ID)
	}

	// Batch fetch rich data
	recipients, _ := s.repo.GetRecipientsByChatIDs(chatIDs, userID)
	lastMessages, _ := s.repo.GetLastMessagesByChatIDs(chatIDs)
	unreadCounts, _ := s.repo.GetUnreadCountsByChatIDs(chatIDs, userID)

	var richChats []ChatDTO
	for _, chat := range chats {
		recipientID := recipients[chat.ID]
		richChats = append(richChats, ChatDTO{
			ID:          chat.ID,
			RecipientID: recipientID,
			LastMessage: lastMessages[chat.ID],
			UnreadCount: unreadCounts[chat.ID],
			IsOnline:    s.hub.IsUserOnline(recipientID),
		})
	}

	return richChats, nil
}
