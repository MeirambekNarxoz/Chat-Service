package services

import (
	"chat-service/internal/models"
	"chat-service/internal/repository"
)

type ChatService struct {
	repo *repository.ChatRepository
}

func NewChatService(repo *repository.ChatRepository) *ChatService {
	return &ChatService{repo: repo}
}

func (s *ChatService) CreatePersonalChat(user1ID, user2ID uint) (*models.Chat, error) {
	chat := &models.Chat{
		Type: models.ChatTypePersonal,
	}
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

func (s *ChatService) GetUserChats(userID uint) ([]models.Chat, error) {
	return s.repo.GetUserChats(userID)
}
func (s *ChatService) CreateGroupChat(name string, userIDs []uint) (*models.Chat, error) {
	chat := &models.Chat{
		Name: name,
		Type: models.ChatTypeGroup,
	}
	if err := s.repo.CreateChat(chat); err != nil {
		return nil, err
	}

	for _, userID := range userIDs {
		_ = s.repo.AddParticipant(&models.ChatParticipant{ChatID: chat.ID, UserID: userID})
	}

	return chat, nil
}
