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

func CreateCategory(c *gin.Context) {
	var body requests.CreateCategoryRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	category := models.Category{
		Name: body.Name,
	}

	if err := database.DB.Preload("Status").Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	if err := database.DB.Preload("Status").First(&category, category.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category created successfully",
		"category": gin.H{
			"id":     category.ID,
			"name":   category.Name,
			"status": category.Status.Name,
		},
	})
}

func CategoryPaginate(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var category []models.Category

	if err := database.DB.Preload("Status").Offset(offset).Limit(limit).Find(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	type CategoryResponse struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	}

	var categoryResponse []CategoryResponse

	for _, cat := range category {
		categoryResponse = append(categoryResponse, CategoryResponse{
			ID:     cat.ID,
			Name:   cat.Name,
			Status: cat.Status.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"categories": gin.H{
			"total": len(categoryResponse),
			"items": categoryResponse,
		},
	})
}

func UpdateCategory(c *gin.Context) {
	var body requests.UpdateCategoryRequest

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request input"})
		return
	}

	category := models.Category{
		ID:       uint(idInt),
		Name:     body.Name,
		StatusID: uint(body.Status),
	}

	if err := database.DB.Preload("Status").Model(&category).Updates(category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully",
		"category": gin.H{
			"id":     category.ID,
			"name":   category.Name,
			"status": category.Status.Name,
		},
	})
}

func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	category := models.Category{
		ID: uint(idInt),
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}
