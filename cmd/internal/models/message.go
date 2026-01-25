package models

import "gorm.io/gorm"

type Conversation struct {
	gorm.Model
	ID               uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ParticipantOneID uint      `gorm:"not null" json:"participant_one_id"`
	ParticipantTwoID uint      `gorm:"not null" json:"participant_two_id"`
	ProductID        uint      `json:"product_id"`
	Product          Product   `json:"product"`
	Messages         []Message `json:"messages"`
}

type Message struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID uint   `gorm:"not null" json:"conversation_id"`
	SenderID       uint   `gorm:"not null" json:"sender_id"`
	Content        string `gorm:"not null" json:"content"`
	IsRead         bool   `gorm:"default:false" json:"is_read"`
}
