package handlers

import (
	"net/http"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/gin-gonic/gin"
)

func GetRoles(c *gin.Context) {
	var roles []models.Role
	var total int64

	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if err := database.DB.Order("id asc").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	type RolesResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	roleResponse := make([]RolesResponse, 0)
	for _, role := range roles {
		roleResponse = append(roleResponse, RolesResponse{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Roles retrieved successfully",
		"roles": gin.H{
			"total": total,
			"items": roleResponse,
		},
	})
}
