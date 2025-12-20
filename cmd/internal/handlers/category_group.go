package handlers

import (
	"net/http"
	"strconv"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/requests"
	"github.com/gin-gonic/gin"
)

func CreateCategoryGroup(c *gin.Context) {
	body := requests.CategoryGroupCreateRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoryGroup := models.CategoryGroup{
		Name: body.Name,
	}

	if err := database.DB.Create(&categoryGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category group created successfully",
		"category_group": gin.H{
			"id":   categoryGroup.ID,
			"name": categoryGroup.Name,
		},
	})
}

func GetAllCategoryGroups(c *gin.Context) {
	var categoryGroups []models.CategoryGroup
	if err := database.DB.Find(&categoryGroups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	type CategoryGroupResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	categoryGroupResponses := []CategoryGroupResponse{}

	for _, categoryGroup := range categoryGroups {
		categoryGroupResponses = append(categoryGroupResponses, CategoryGroupResponse{
			ID:   categoryGroup.ID,
			Name: categoryGroup.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Category groups retrieved successfully",
		"category_groups": categoryGroupResponses,
	})
}

func UpdateCategoryGroup(c *gin.Context) {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	body := requests.CategoryGroupCreateRequest{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoryGroup := models.CategoryGroup{
		ID:   uint(idInt),
		Name: body.Name,
	}

	if err := database.DB.Model(&categoryGroup).Updates(categoryGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category group updated successfully",
		"category_group": gin.H{
			"id":   categoryGroup.ID,
			"name": categoryGroup.Name,
		}})
}

func DeleteCategoryGroup(c *gin.Context) {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	categoryGroup := models.CategoryGroup{
		ID: uint(idInt),
	}

	if err := database.DB.Delete(&categoryGroup).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category group deleted successfully"})
}
