package services

import (
	"fmt"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

// CreateNotification creates a notification (personal if userID is provided, system-wide if nil)
// Read status is tracked in NotificationRead table for each user
func CreateNotification(userID *uint, title, message, notificationType string, reference *string) (*models.Notification, error) {
	notification := models.Notification{
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      notificationType,
		Reference: reference,
	}

	if err := database.DB.Create(&notification).Error; err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &notification, nil
}

// NotifyAdmins creates a system-wide notification (user_id = nil) visible to all users
// Read status for each user is tracked in NotificationRead table
func NotifyAdmins(title, message, notificationType string, reference *string) error {
	// Create system-wide notification (user_id = nil) - visible to all users
	_, err := CreateNotification(nil, title, message, notificationType, reference)
	return err
}

// NotifyUser creates a personal notification (user_id set) for a specific user
// Read status is tracked in NotificationRead table
func NotifyUser(userID uint, title, message, notificationType string, reference *string) error {
	userIDPtr := &userID
	_, err := CreateNotification(userIDPtr, title, message, notificationType, reference)
	return err
}
