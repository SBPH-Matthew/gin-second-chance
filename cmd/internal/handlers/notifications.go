package handlers

import (
	"net/http"
	"strconv"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/gin-gonic/gin"
)

func GetNotifications(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	// Get notifications: personal notifications for this user OR system-wide notifications (user_id is NULL)
	var notifications []models.Notification
	if err := database.DB.Where("user_id = ? OR user_id IS NULL", userID).
		Order("created_at DESC").
		Limit(50).
		Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	// Get read status for this user
	var readNotificationIDs []uint
	database.DB.Model(&models.NotificationRead{}).
		Where("user_id = ? AND is_read = ?", userID, true).
		Pluck("notification_id", &readNotificationIDs)

	// Create a map for quick lookup
	readMap := make(map[uint]bool)
	for _, id := range readNotificationIDs {
		readMap[id] = true
	}

	// Add is_read field to each notification based on user's read status
	type NotificationWithReadStatus struct {
		models.Notification
		IsRead bool `json:"is_read"`
	}

	notificationsWithStatus := make([]NotificationWithReadStatus, len(notifications))
	for i, notif := range notifications {
		notificationsWithStatus[i] = NotificationWithReadStatus{
			Notification: notif,
			IsRead:       readMap[notif.ID],
		}
	}

	// Count unread notifications
	unreadCount := int64(len(notifications) - len(readNotificationIDs))

	c.JSON(http.StatusOK, gin.H{
		"message":       "Notifications retrieved successfully",
		"notifications": notificationsWithStatus,
		"unread_count":  unreadCount,
	})
}

func MarkNotificationAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid notification ID"})
		return
	}

	// Check if notification exists and user has access (personal notification for this user OR system-wide)
	var notification models.Notification
	if err := database.DB.Where("id = ? AND (user_id = ? OR user_id IS NULL)", idInt, userID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Notification not found or access denied"})
		return
	}

	// Create or update NotificationRead record
	var notificationRead models.NotificationRead
	result := database.DB.Where("notification_id = ? AND user_id = ?", idInt, userID).First(&notificationRead)

	if result.Error != nil {
		// Create new read record
		notificationRead = models.NotificationRead{
			NotificationID: uint(idInt),
			UserID:         userID,
			IsRead:         true,
		}
		if err := database.DB.Create(&notificationRead).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
			return
		}
	} else {
		// Update existing record
		notificationRead.IsRead = true
		if err := database.DB.Save(&notificationRead).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

func MarkAllNotificationsAsRead(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	// Get all notifications accessible to this user (personal OR system-wide)
	var allNotifications []models.Notification
	if err := database.DB.Where("user_id = ? OR user_id IS NULL", userID).Find(&allNotifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	var unreadNotifications []models.Notification

	var readNotificationIDs []uint
	database.DB.Model(&models.NotificationRead{}).
		Where("user_id = ? AND is_read = ?", userID, true).
		Pluck("notification_id", &readNotificationIDs)

	readMap := make(map[uint]bool)
	for _, id := range readNotificationIDs {
		readMap[id] = true
	}

	// Find unread notifications
	for _, notif := range allNotifications {
		if !readMap[notif.ID] {
			unreadNotifications = append(unreadNotifications, notif)
		}
	}

	// Mark all unread notifications as read
	for _, notif := range unreadNotifications {
		var notificationRead models.NotificationRead
		result := database.DB.Where("notification_id = ? AND user_id = ?", notif.ID, userID).First(&notificationRead)

		if result.Error != nil {
			// Create new read record
			notificationRead = models.NotificationRead{
				NotificationID: notif.ID,
				UserID:         userID,
				IsRead:         true,
			}
			database.DB.Create(&notificationRead)
		} else {
			// Update existing record
			notificationRead.IsRead = true
			database.DB.Save(&notificationRead)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}
