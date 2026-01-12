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

func CreateProduct(c *gin.Context) {
	var body requests.CreateProductRequest

	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := utils.Validate.Struct(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	product := models.Product{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		SellerID:    user.ID,
	}

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product created successfully",
		"product": gin.H{
			"id":          product.ID,
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"user": gin.H{
				"id":         user.ID,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
				"email":      user.Email,
			},
		},
	})
}

func UpdateProduct(c *gin.Context) {
	var body requests.CreateProductRequest

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product ID"})
		return
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if err := utils.Validate.Struct(body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	// Check if product exists and belongs to the user
	var existingProduct models.Product
	if err := database.DB.Where("id = ? AND seller_id = ?", idInt, userID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found or access denied"})
		return
	}

	// Update fields
	existingProduct.Name = body.Name
	existingProduct.Description = body.Description
	existingProduct.Price = body.Price

	if err := database.DB.Save(&existingProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
	})
}

func ProductPaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")

	if page == "" || limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing page or limit query parameter"})
		return
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page query parameter"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid limit query parameter"})
		return
	}

	offset := (pageInt - 1) * limitInt

	var products []models.Product
	var total int64

	query := database.DB.Model(&models.Product{})

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Get paginated items
	if err := query.Preload("Status").Preload("ProductCondition").Preload("Category").Offset(offset).Limit(limitInt).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	type ProductResponse struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Status      string  `json:"status"`
		Condition   string  `json:"condition"`
		Category    string  `json:"category"`
	}

	var productResponses []ProductResponse

	for _, product := range products {
		productResponses = append(productResponses, ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Status:      product.Status.Name,
			Condition:   product.ProductCondition.Name,
			Category:    product.Category.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products paginated successfully",
		"products": gin.H{
			"total": total,
			"items": productResponses,
		},
	})
}

func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product ID"})
		return
	}

	userID := c.GetUint("user_id")

	// Only allow deletion if the user owns the product
	result := database.DB.Where("id = ? AND seller_id = ?", idInt, userID).Delete(&models.Product{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + result.Error.Error(),
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

func ProductDetails(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := database.DB.Preload("Status").First(&product, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product details retrieved successfully",
		"product": gin.H{
			"id":          product.ID,
			"name":        product.Name,
			"description": product.Description,
			"price":       product.Price,
			"status":      product.Status,
		},
	})
}

func GetMyProductsPaginate(c *gin.Context) {
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

	userID := c.GetUint("user_id")

	offset := (page - 1) * limit

	var products []models.Product
	var total int64

	query := database.DB.Model(&models.Product{}).Where("seller_id = ?", userID)

	// Get total count for this user
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	if err := query.Preload("Status").Preload("ProductCondition").Preload("Category").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	type ProductResponse struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Status      string  `json:"status"`
		Condition   string  `json:"condition"`
		Category    string  `json:"category"`
	}

	var productResponses []ProductResponse

	for _, product := range products {
		productResponses = append(productResponses, ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Status:      product.Status.Name,
			Condition:   product.ProductCondition.Name,
			Category:    product.Category.Name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "My products retrieved successfully",
		"products": gin.H{
			"total": total,
			"items": productResponses,
		},
	})
}
