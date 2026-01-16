package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	ID        uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    *uint   `json:"user_id"` // Optional: if set, personal notification for this user; if nil, system-wide notification for all users
	Title     string  `json:"title"`
	Message   string  `json:"message"`
	Type      string  `json:"type"`      // info, success, warning, error
	Reference *string `json:"reference"` // Optional: reference to related entity (e.g., "product:123", "user:456")

	User User `gorm:"foreignKey:UserID" json:"user"`
}

// NotificationRead tracks which users have read which notifications
type NotificationRead struct {
	gorm.Model
	ID             uint `gorm:"primaryKey;autoIncrement" json:"id"`
	NotificationID uint `json:"notification_id"`
	UserID         uint `json:"user_id"`
	IsRead         bool `json:"is_read" gorm:"default:true"`

	Notification Notification `gorm:"foreignKey:NotificationID" json:"notification"`
	User         User         `gorm:"foreignKey:UserID" json:"user"`
}
