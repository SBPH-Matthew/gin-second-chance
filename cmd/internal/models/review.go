package models

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	ID              uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	ConversationID  uint        `gorm:"not null;uniqueIndex:idx_reviewer_conversation" json:"conversation_id"`
	ReviewerID      uint        `gorm:"not null;uniqueIndex:idx_reviewer_conversation" json:"reviewer_id"`
	Reviewer        User        `gorm:"foreignKey:ReviewerID" json:"reviewer"`
	TargetUserID    uint        `gorm:"not null" json:"target_user_id"`
	TargetUser      User        `gorm:"foreignKey:TargetUserID" json:"target_user"`
	ProductID       uint        `json:"product_id"`
	Product         Product     `json:"product"`
	Rating          int         `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Tags            StringArray `gorm:"type:json" json:"tags"`           // structured tags
	Comment         string      `json:"comment"`                         // public comment
	InternalComment string      `json:"internal_comment"`                // for admin/moderation
	TrustWeight     float64     `gorm:"default:1.0" json:"trust_weight"` // for future weighting
}
