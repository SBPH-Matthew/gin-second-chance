package models

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	ID           uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	ReviewerID   uint    `gorm:"not null" json:"reviewer_id"`
	Reviewer     User    `gorm:"foreignKey:ReviewerID" json:"reviewer"`
	TargetUserID uint    `gorm:"not null" json:"target_user_id"`
	TargetUser   User    `gorm:"foreignKey:TargetUserID" json:"target_user"`
	ProductID    uint    `json:"product_id"`
	Product      Product `json:"product"`
	Rating       int     `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment      string  `json:"comment"`
}
