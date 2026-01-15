package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/requests"
	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse multipart form: " + err.Error()})
		return
	}

	// Get form values
	name := c.PostForm("name")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")
	status := c.PostForm("status")
	condition := c.PostForm("condition")
	category := c.PostForm("category")

	// Validate required fields
	if name == "" || priceStr == "" || status == "" || condition == "" || category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing required fields"})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid price"})
		return
	}

	intStatus, err := strconv.Atoi(status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status ID"})
		return
	}

	intCondition, err := strconv.Atoi(condition)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid condition ID",
		})
		return
	}

	intCategory, err := strconv.Atoi(category)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid category ID",
		})
		return
	}

	// Handle file uploads
	var imagePaths models.StringArray
	multipartForm, err := c.MultipartForm()
	if err == nil && multipartForm != nil {
		formFiles := multipartForm.File["images"]

		// Create uploads directory if it doesn't exist
		uploadDir := "./uploads/products"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to create upload directory: " + err.Error(),
			})
			return
		}

		// Process uploaded files
		for _, fileHeader := range formFiles {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			// Generate unique filename
			timestamp := time.Now().UnixNano()
			// Get the base filename without path
			originalFilename := filepath.Base(fileHeader.Filename)
			// Get extension (includes the dot, e.g., ".jpg")
			ext := filepath.Ext(originalFilename)
			if ext == "" {
				// If no extension found, try to detect from content type or use default
				ext = ".jpg"
			}
			// Remove extension from original filename (handle potential duplicates)
			// Keep removing the extension until it's gone
			baseName := originalFilename
			extLower := strings.ToLower(ext)
			for {
				baseLower := strings.ToLower(baseName)
				if strings.HasSuffix(baseLower, extLower) {
					baseName = baseName[:len(baseName)-len(ext)]
				} else {
					break
				}
			}
			// Clean the base name (remove spaces, special chars)
			cleanBaseName := strings.ReplaceAll(baseName, " ", "_")
			// Ensure extension is lowercase and add it back once
			ext = extLower
			filename := fmt.Sprintf("%d_%s%s", timestamp, cleanBaseName, ext)
			filePath := filepath.Join(uploadDir, filename)

			// Create the file
			dst, err := os.Create(filePath)
			if err != nil {
				file.Close()
				continue
			}

			// Copy file content
			if _, err := io.Copy(dst, file); err != nil {
				file.Close()
				dst.Close()
				os.Remove(filePath) // Clean up on error
				continue
			}

			file.Close()
			dst.Close()

			// Store relative path
			imagePaths = append(imagePaths, fmt.Sprintf("/uploads/products/%s", filename))
		}
	}

	product := models.Product{
		Name:               name,
		Description:        description,
		Price:              price,
		Images:             imagePaths,
		StatusID:           uint(intStatus),
		SellerID:           user.ID,
		ProductConditionID: uint(intCondition),
		CategoryID:         uint(intCategory),
	}

	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product created successfully",
	})
}

func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product ID"})
		return
	}

	userID := c.GetUint("user_id")

	// Check if product exists and belongs to the user
	var existingProduct models.Product
	if err := database.DB.Where("id = ? AND seller_id = ?", idInt, userID).First(&existingProduct).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found or access denied"})
		return
	}

	// Parse multipart form if present, otherwise try JSON
	if c.ContentType() == "application/json" || c.ContentType() == "" {
		var body requests.CreateProductRequest
		if err := c.ShouldBindJSON(&body); err == nil {
			intStatus, _ := strconv.Atoi(body.Status)
			intCondition, _ := strconv.Atoi(body.Condition)
			intCategory, _ := strconv.Atoi(body.Category)

			existingProduct.Name = body.Name
			existingProduct.Description = body.Description
			existingProduct.Price = body.Price
			existingProduct.StatusID = uint(intStatus)
			existingProduct.ProductConditionID = uint(intCondition)
			existingProduct.CategoryID = uint(intCategory)

			if err := database.DB.Save(&existingProduct).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Database error: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Product updated successfully",
			})
			return
		}
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse multipart form: " + err.Error()})
		return
	}

	// Get form values
	if name := c.PostForm("name"); name != "" {
		existingProduct.Name = name
	}
	if description := c.PostForm("description"); description != "" {
		existingProduct.Description = description
	}
	if priceStr := c.PostForm("price"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			existingProduct.Price = price
		}
	}
	if status := c.PostForm("status"); status != "" {
		if intStatus, err := strconv.Atoi(status); err == nil {
			existingProduct.StatusID = uint(intStatus)
		}
	}
	if condition := c.PostForm("condition"); condition != "" {
		if intCondition, err := strconv.Atoi(condition); err == nil {
			existingProduct.ProductConditionID = uint(intCondition)
		}
	}
	if category := c.PostForm("category"); category != "" {
		if intCategory, err := strconv.Atoi(category); err == nil {
			existingProduct.CategoryID = uint(intCategory)
		}
	}

	// Handle existing images to keep (from form data)
	var imagesToKeep models.StringArray
	existingImagesJSON := c.PostForm("existingImages")
	if existingImagesJSON != "" {
		// Parse JSON array of image paths
		var existingImagesList []string
		if err := json.Unmarshal([]byte(existingImagesJSON), &existingImagesList); err == nil {
			imagesToKeep = models.StringArray(existingImagesList)
		}
	} else {
		// If existingImages field is not provided, keep all existing images
		imagesToKeep = existingProduct.Images
	}

	// Delete files that are no longer in the images to keep
	for _, oldImagePath := range existingProduct.Images {
		shouldKeep := false
		for _, keepPath := range imagesToKeep {
			if oldImagePath == keepPath {
				shouldKeep = true
				break
			}
		}
		if !shouldKeep {
			// Remove the file from storage
			// oldImagePath is like "/uploads/products/filename.jpg"
			// Convert to "./uploads/products/filename.jpg"
			fullPath := fmt.Sprintf(".%s", oldImagePath)
			if err := os.Remove(fullPath); err != nil {
				// Log error but don't fail the update
				fmt.Printf("Failed to delete image file %s: %v\n", fullPath, err)
			}
		}
	}

	// Handle new file uploads
	multipartForm, err := c.MultipartForm()
	var newImagePaths models.StringArray
	if err == nil && multipartForm != nil {
		formFiles := multipartForm.File["images"]
		if len(formFiles) > 0 {
			uploadDir := "./uploads/products"
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Failed to create upload directory: " + err.Error(),
				})
				return
			}

			// Process uploaded files
			for _, fileHeader := range formFiles {
				file, err := fileHeader.Open()
				if err != nil {
					continue
				}

				// Generate unique filename
				timestamp := time.Now().UnixNano()
				// Get the base filename without path
				originalFilename := filepath.Base(fileHeader.Filename)
				// Get extension (includes the dot, e.g., ".jpg")
				ext := filepath.Ext(originalFilename)
				if ext == "" {
					// If no extension found, try to detect from content type or use default
					ext = ".jpg"
				}
				// Remove extension from original filename (handle potential duplicates)
				// Keep removing the extension until it's gone
				baseName := originalFilename
				extLower := strings.ToLower(ext)
				for {
					baseLower := strings.ToLower(baseName)
					if strings.HasSuffix(baseLower, extLower) {
						baseName = baseName[:len(baseName)-len(ext)]
					} else {
						break
					}
				}
				// Clean the base name (remove spaces, special chars)
				cleanBaseName := strings.ReplaceAll(baseName, " ", "_")
				// Ensure extension is lowercase and add it back once
				ext = extLower
				filename := fmt.Sprintf("%d_%s%s", timestamp, cleanBaseName, ext)
				filePath := filepath.Join(uploadDir, filename)

				// Create the file
				dst, err := os.Create(filePath)
				if err != nil {
					file.Close()
					continue
				}

				// Copy file content
				if _, err := io.Copy(dst, file); err != nil {
					file.Close()
					dst.Close()
					os.Remove(filePath) // Clean up on error
					continue
				}

				file.Close()
				dst.Close()

				// Add to new images
				newImagePaths = append(newImagePaths, fmt.Sprintf("/uploads/products/%s", filename))
			}
		}
	}

	// Update images: keep existing ones that weren't deleted + add new ones
	// If existingImagesJSON was provided (even if empty), use it; otherwise keep existing
	if existingImagesJSON != "" {
		// existingImages field was provided, use it (could be empty array to delete all)
		existingProduct.Images = append(imagesToKeep, newImagePaths...)
	} else if len(newImagePaths) > 0 {
		// No existingImages field but new images added, append to existing
		existingProduct.Images = append(existingProduct.Images, newImagePaths...)
	}
	// If neither existingImages nor new images, keep existing images unchanged

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
		ID          uint                    `json:"id"`
		Name        string                  `json:"name"`
		Description string                  `json:"description"`
		Price       float64                 `json:"price"`
		Images      models.StringArray      `json:"images"`
		Status      models.ProductStatus    `json:"status"`
		Condition   models.ProductCondition `json:"condition"`
		Category    models.Category         `json:"category"`
	}

	var productResponses []ProductResponse

	for _, product := range products {
		productResponses = append(productResponses, ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Images:      product.Images,
			Status:      product.Status,
			Condition:   product.ProductCondition,
			Category:    product.Category,
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
