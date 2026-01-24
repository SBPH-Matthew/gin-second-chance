package handlers

import (
	"net/http"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/gin-gonic/gin"
)

func Profile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.Preload("Role").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Authenticated request",
		"user":    user,
	})
}

func Logout(c *gin.Context) {
	// Clear the authentication cookie
	c.SetCookie(
		"token",     // Name of the cookie
		"",          // Empty value
		-1,          // Negative MaxAge to expire immediately
		"/",         // Path
		"localhost", // Domain (change to your actual domain in production)
		false,       // Secure
		true,        // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}
