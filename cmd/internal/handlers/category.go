package handlers

import (
	"net/http"
	"strconv"
	"strings"

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
		Name:            body.Name,
		CategoryGroupID: body.CategoryGroup,
		StatusID:        body.Status,
	}

	if err := database.DB.Where("name = ?", category.Name).First(&models.Category{}).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Category name already exists",
			"errors": gin.H{
				"name": "Category name already exists",
			},
		})
		return
	}

	if err := database.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	if err := database.DB.Preload("Status").First(&category, category.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
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

	var categories []models.Category
	var total int64 // Use int64 for GORM Count

	// 1. Get the TOTAL count of all records (without offset/limit)
	if err := database.DB.Model(&models.Category{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	// 2. Fetch the specific page items
	if err := database.DB.Preload("Status").Preload("CategoryGroup").
		Order("id asc").
		Offset(offset).
		Limit(limit).
		Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	// 3. Map to Response struct
	type CategoryResponse struct {
		ID     uint   `json:"id"`
		Name   string `json:"name"`
		Status uint   `json:"status"`
		Group  uint   `json:"category_group"`
	}

	categoryResponse := make([]CategoryResponse, 0)
	for _, cat := range categories {
		categoryResponse = append(categoryResponse, CategoryResponse{
			ID:     cat.ID,
			Name:   cat.Name,
			Status: cat.StatusID,
			Group:  cat.CategoryGroupID,
		})
	}

	// 4. Return both the items and the actual total
	c.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"categories": gin.H{
			"total": total, // This is now the global total (e.g., 100)
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
			"message": "Invalid category ID",
			"errors": gin.H{
				"name": "Invalid category ID",
			},
		})
		return
	}

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid form data",
			"errors": gin.H{
				"name": "Invalid Form",
			}})
		return
	}

	body.Name = strings.ToLower(body.Name)

	category := models.Category{
		ID:              uint(idInt),
		Name:            body.Name,
		StatusID:        uint(body.Status),
		CategoryGroupID: uint(body.CategoryGroup),
	}

	if err := database.DB.Model(&category).Updates(category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	if err := database.DB.Preload("Status").Preload("CategoryGroup").First(&category, category.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully",
		"category": gin.H{
			"id":             category.ID,
			"name":           category.Name,
			"status":         category.Status.Name,
			"category_group": category.CategoryGroup.Name,
		},
	})
}

func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{

			"message": "Invalid category ID",
		})
		return
	}

	category := models.Category{
		ID: uint(idInt),
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}
