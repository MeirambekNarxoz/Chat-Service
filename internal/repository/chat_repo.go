package repository

import (
	"chat-service/internal/models"
	"gorm.io/gorm"
)

type ChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateChat(chat *models.Chat) error {
	return r.db.Create(chat).Error
}

func (r *ChatRepository) AddParticipant(chatParticipant *models.ChatParticipant) error {
	return r.db.Create(chatParticipant).Error
}

func (r *ChatRepository) SaveMessage(msg *models.Message) error {
	return r.db.Create(msg).Error
}

func (r *ChatRepository) GetMessagesByChatID(chatID uint) ([]models.Message, error) {
	var messages []models.Message
	err := r.db.Where("chat_id = ?", chatID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *ChatRepository) GetUserChats(userID uint) ([]models.Chat, error) {
	var chats []models.Chat
	err := r.db.Joins("JOIN chat_participants ON chat_participants.chat_id = chats.id").
		Where("chat_participants.user_id = ?", userID).
		Find(&chats).Error
	return chats, err
}
func (r *ChatRepository) GetChatParticipants(chatID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&models.ChatParticipant{}).Where("chat_id = ?", chatID).Pluck("user_id", &userIDs).Error
	return userIDs, err
}
