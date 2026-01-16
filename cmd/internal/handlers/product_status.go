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

func CreateProductStatus(c *gin.Context) {
	var body requests.CreateProductStatusRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	productStatus := models.ProductStatus{
		Name: body.Name,
	}

	if err := database.DB.Create(&productStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product status created successfully",
	})
}

func UpdateProductStatus(c *gin.Context) {
	var body requests.CreateProductStatusRequest

	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	productStatus := models.ProductStatus{
		ID:   uint(idInt),
		Name: body.Name,
	}

	if err := database.DB.Model(&productStatus).Updates(productStatus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product status updated successfully",
	})
}

func DeleteProductStatus(c *gin.Context) {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := database.DB.Delete(&models.ProductStatus{}, idInt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product status deleted successfully",
	})
}

func GetAllProductStatus(c *gin.Context) {
	var statuses []models.ProductStatus

	if err := database.DB.Find(&statuses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Product statuses retrieved successfully",
		"product_status": statuses,
	})
}
