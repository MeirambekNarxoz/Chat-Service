package models

import "time"

type ChatType string

const (
	ChatTypePersonal ChatType = "PERSONAL"
	ChatTypeGroup    ChatType = "GROUP"
)

type Chat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(255)" json:"name"` // Used mainly for GROUP chats
	Type      ChatType  `gorm:"type:varchar(20);not null" json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatParticipant struct {
	ChatID   uint      `gorm:"primaryKey" json:"chat_id"`
	UserID   uint      `gorm:"primaryKey" json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ChatID    uint      `gorm:"not null;index" json:"chat_id"`
	SenderID  uint      `gorm:"not null" json:"sender_id"`
	Text      string    `gorm:"type:text" json:"text"`
	FileURL   string    `gorm:"type:text" json:"file_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
