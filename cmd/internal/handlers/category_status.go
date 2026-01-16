package handlers

import (
	"net/http"
	"strconv"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/requests"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
	"github.com/gin-gonic/gin"
)

func CreateCategoryStatus(c *gin.Context) {
	var body requests.CreateCategoryStatusRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	categoryStatus := models.CategoryStatus{
		Name: body.Name,
	}

	if err := database.DB.Create(&categoryStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category status created successfully",
	})
}

func GetAllCategoryStatuses(c *gin.Context) {
	categoryStatuses := []models.CategoryStatus{}

	if err := database.DB.Where("name != ?", "ARCHIVED").Find(&categoryStatuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Category statuses retrieved successfully",
		"category_statuses": categoryStatuses,
	})
}

func UpdateCategoryStatus(c *gin.Context) {
	var body requests.CreateCategoryStatusRequest

	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	categoryStatus := models.CategoryStatus{
		ID:   uint(idInt),
		Name: body.Name,
	}

	if err := database.DB.Model(&categoryStatus).Updates(categoryStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category status updated successfully",
	})
}

func DeleteCategoryStatus(c *gin.Context) {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Delete(&models.CategoryStatus{}, idInt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category status deleted successfully",
	})
}
