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

func CreateProductCondition(c *gin.Context) {
	var body requests.CreateProductConditionRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	productCondition := models.ProductCondition{
		Name: body.Name,
	}

	if err := database.DB.Create(&productCondition).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product condition created successfully",
		"product_condition": gin.H{
			"id":   productCondition.ID,
			"name": productCondition.Name,
		},
	})
}

func GetAllProductCondition(c *gin.Context) {
	var productConditions []models.ProductCondition

	if err := database.DB.Find(&productConditions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	type ProductConditionResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	var productConditionResponses []ProductConditionResponse

	for _, productCondition := range productConditions {
		productConditionResponses = append(productConditionResponses, ProductConditionResponse{
			ID:   productCondition.ID,
			Name: productCondition.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "Product conditions retrieved successfully",
		"product_conditions": productConditionResponses,
	})
}

func UpdateProductCondition(c *gin.Context) {
	var body requests.CreateProductConditionRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	idStr := c.Param("id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	productCondition := models.ProductCondition{
		ID:   uint(idInt),
		Name: body.Name,
	}

	if err := database.DB.Model(&productCondition).Updates(productCondition).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product condition updated successfully",
		"product_condition": gin.H{
			"id":   productCondition.ID,
			"name": productCondition.Name,
		},
	})
}

func DeleteProductCondition(c *gin.Context) {
	idStr := c.Param("id")
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID",
		})
		return
	}

	if err := database.DB.Delete(&models.ProductCondition{}, idInt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product condition deleted successfully",
	})
}
