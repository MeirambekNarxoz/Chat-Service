package repository

import (
	"chat-service/internal/models"
	"time"

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
	if chatParticipant.JoinedAt.IsZero() {
		chatParticipant.JoinedAt = time.Now()
	}
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
		Order("chats.created_at desc").
		Find(&chats).Error
	return chats, err
}

func (r *ChatRepository) GetRecipientID(chatID, userID uint) (uint, error) {
	var participant models.ChatParticipant
	err := r.db.Where("chat_id = ? AND user_id != ?", chatID, userID).First(&participant).Error
	return participant.UserID, err
}

func (r *ChatRepository) GetLastMessage(chatID uint) (*models.Message, error) {
	var msg models.Message
	err := r.db.Where("chat_id = ?", chatID).Order("created_at desc").First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *ChatRepository) GetUnreadCount(chatID, userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Message{}).
		Where("chat_id = ? AND sender_id != ? AND is_read = ?", chatID, userID, false).
		Count(&count).Error
	return count, err
}

func (r *ChatRepository) MarkAsRead(chatID, userID uint) error {
	return r.db.Model(&models.Message{}).
		Where("chat_id = ? AND sender_id != ? AND is_read = ?", chatID, userID, false).
		Update("is_read", true).Error
}

func (r *ChatRepository) GetChatParticipants(chatID uint) ([]uint, error) {
	var userIDs []uint
	err := r.db.Model(&models.ChatParticipant{}).Where("chat_id = ?", chatID).Pluck("user_id", &userIDs).Error
	return userIDs, err
}

func (r *ChatRepository) IsChatParticipant(chatID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.ChatParticipant{}).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ChatRepository) GetPersonalChat(user1ID, user2ID uint) (*models.Chat, error) {
	var _ models.ChatParticipant
	// Находим чаты, в которых участвует первый пользователь
	// и проверяем, участвует ли в них же второй пользователь
	query := `
		SELECT cp1.chat_id 
		FROM chat_participants cp1
		JOIN chat_participants cp2 ON cp1.chat_id = cp2.chat_id
		JOIN chats c ON c.id = cp1.chat_id
		WHERE cp1.user_id = ? AND cp2.user_id = ?
		LIMIT 1
	`
	var chatID uint
	err := r.db.Raw(query, user1ID, user2ID).Scan(&chatID).Error
	if err != nil || chatID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var chat models.Chat
	err = r.db.First(&chat, chatID).Error
	return &chat, err
}

func (r *ChatRepository) GetRecipientsByChatIDs(chatIDs []uint, userID uint) (map[uint]uint, error) {
	var participants []models.ChatParticipant
	err := r.db.Where("chat_id IN ? AND user_id != ?", chatIDs, userID).Find(&participants).Error
	if err != nil {
		return nil, err
	}

	res := make(map[uint]uint)
	for _, p := range participants {
		res[p.ChatID] = p.UserID
	}
	return res, nil
}

func (r *ChatRepository) GetLastMessagesByChatIDs(chatIDs []uint) (map[uint]*models.Message, error) {
	if len(chatIDs) == 0 {
		return map[uint]*models.Message{}, nil
	}

	var messages []models.Message
	// PostgreSQL: latest message per chat_id
	err := r.db.Raw(`
		SELECT DISTINCT ON (chat_id) id, chat_id, sender_id, text, file_url, is_read, type, created_at
		FROM messages
		WHERE chat_id IN ?
		ORDER BY chat_id, created_at DESC
	`, chatIDs).Scan(&messages).Error
	if err != nil {
		return nil, err
	}

	res := make(map[uint]*models.Message)
	for i := range messages {
		res[messages[i].ChatID] = &messages[i]
	}
	return res, nil
}

func (r *ChatRepository) GetUnreadCountsByChatIDs(chatIDs []uint, userID uint) (map[uint]int64, error) {
	type Result struct {
		ChatID uint
		Count  int64
	}
	var results []Result
	err := r.db.Model(&models.Message{}).
		Select("chat_id, count(*) as count").
		Where("chat_id IN ? AND sender_id != ? AND is_read = ?", chatIDs, userID, false).
		Group("chat_id").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	res := make(map[uint]int64)
	for _, r := range results {
		res[r.ChatID] = r.Count
	}
	return res, nil
}
