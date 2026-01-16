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
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
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
	})
}

func CategoryPaginate(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := c.Query("search")

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
	var total int64

	// Build base query for counting
	baseQuery := database.DB.Model(&models.Category{})

	// Apply search filter if provided
	if search != "" && strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		baseQuery = baseQuery.Joins("LEFT JOIN category_statuses ON categories.status_id = category_statuses.id").
			Joins("LEFT JOIN category_groups ON categories.category_group_id = category_groups.id").
			Where(
				"categories.name ILIKE ? OR category_statuses.name ILIKE ? OR category_groups.name ILIKE ?",
				searchTerm, searchTerm, searchTerm,
			).
			Group("categories.id") // Group by to avoid duplicates from joins
	}

	// Get total count with search filter applied
	if err := baseQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	// Build query for fetching categories
	// Use a subquery approach to get matching IDs first, then fetch with Preload
	if search != "" && strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		var categoryIDs []uint
		if err := database.DB.Model(&models.Category{}).
			Joins("LEFT JOIN category_statuses ON categories.status_id = category_statuses.id").
			Joins("LEFT JOIN category_groups ON categories.category_group_id = category_groups.id").
			Where(
				"categories.name ILIKE ? OR category_statuses.name ILIKE ? OR category_groups.name ILIKE ?",
				searchTerm, searchTerm, searchTerm,
			).
			Group("categories.id").
			Pluck("categories.id", &categoryIDs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
			return
		}

		// If no matching IDs, return empty result
		if len(categoryIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "Categories retrieved successfully",
				"categories": gin.H{
					"total": 0,
					"items": []models.Category{},
				},
			})
			return
		}

		// Fetch categories by IDs with Preload
		if err := database.DB.Preload("Status").Preload("CategoryGroup").
			Where("categories.id IN ?", categoryIDs).
			Order("id asc").
			Offset(offset).
			Limit(limit).
			Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
			return
		}
	} else {
		// No search - fetch all with Preload
		if err := database.DB.Preload("Status").Preload("CategoryGroup").
			Order("id asc").
			Offset(offset).
			Limit(limit).
			Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Categories retrieved successfully",
		"categories": gin.H{
			"total": total,
			"items": categories,
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

	if err := database.DB.Where("name = ?", category.Name).First(&models.Category{}).Error; err == nil {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "Category name already exists",
			"errors": gin.H{
				"name": "Category name already exists",
			},
		})
		return
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

func SetCategoryStatus(c *gin.Context) {
	id := c.Param("id")

	type ChangeStatusRequest struct {
		Status uint `json:"status" validate:"required"`
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	var request = ChangeStatusRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	category := models.Category{
		ID:       uint(idInt),
		StatusID: uint(request.Status),
	}

	if err := database.DB.Model(&category).Updates(category).Updates(category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category status updated successfully",
	})
}

func GetAllCategory(c *gin.Context) {
	var categories []models.Category

	if err := database.DB.Preload("Status").Preload("CategoryGroup").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Categories retrieved successfully",
		"categories": categories,
	})
}
