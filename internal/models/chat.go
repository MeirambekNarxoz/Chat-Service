package models

import "time"

type Chat struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
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
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
