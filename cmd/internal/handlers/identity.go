package handlers

import (
	"net/http"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/gin-gonic/gin"
)

func VerifyIdentity(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// In a real app, you'd process ID documents here.
	// For this simulation, we'll just set it to true.
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("identity_verified", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to verify identity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Identity verified successfully"})
}

func UpdatePhone(c *gin.Context) {
	type PhoneRequest struct {
		Phone string `json:"phone" binding:"required"`
	}

	var body PhoneRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid phone number"})
		return
	}

	userID := c.MustGet("userID").(uint)

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("phone", body.Phone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update phone"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phone number updated"})
}
