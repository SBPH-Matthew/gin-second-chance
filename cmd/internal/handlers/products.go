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
	"github.com/SBPH-Matthew/second-chance/cmd/internal/services"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
	"github.com/gin-gonic/gin"
)

func getUploadDir() string {
	dir := os.Getenv("UPLOAD_DIR")
	if dir == "" {
		dir = "./uploads"
	}
	return dir
}

func CreateProduct(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	// Validate form data
	var req requests.CreateProductRequest
	if err := utils.ValidateBodyFormData(c, &req); err != nil {
		return // Error response already sent by ValidateBodyFormData
	}

	// Price is already parsed by the validator
	price := req.Price

	// Parse IDs
	intStatus, err := strconv.Atoi(req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status ID"})
		return
	}

	intCondition, err := strconv.Atoi(req.Condition)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid condition ID",
		})
		return
	}

	intCategory, err := strconv.Atoi(req.Category)
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
		uploadDir := filepath.Join(getUploadDir(), "products")
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
		Name:               req.Name,
		Description:        req.Description,
		Price:              price,
		Location:           req.Location,
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

	// Create notification for admins about new product
	reference := fmt.Sprintf("product:%d", product.ID)
	services.NotifyAdmins(
		"New Product Created",
		fmt.Sprintf("A new product '%s' has been created by %s %s", product.Name, user.FirstName, user.LastName),
		"info",
		&reference,
	)

	// Notify the seller
	services.NotifyUser(
		user.ID,
		"Product Created Successfully",
		fmt.Sprintf("Your product '%s' has been created successfully", product.Name),
		"success",
		&reference,
	)

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
			existingProduct.Location = body.Location
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
	if location := c.PostForm("location"); location != "" {
		existingProduct.Location = location
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
			// Convert to full path using upload dir
			relativePath := strings.TrimPrefix(oldImagePath, "/uploads")
			fullPath := filepath.Join(getUploadDir(), relativePath)
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
			uploadDir := filepath.Join(getUploadDir(), "products")
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

	var products []models.Product
	var total int64

	// Build base query for counting
	baseQuery := database.DB.Model(&models.Product{})

	// Apply search filter if provided
	if search != "" && strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		baseQuery = baseQuery.Joins("LEFT JOIN categories ON products.category_id = categories.id").
			Joins("LEFT JOIN product_statuses ON products.status_id = product_statuses.id").
			Joins("LEFT JOIN product_conditions ON products.product_condition_id = product_conditions.id").
			Where(
				"products.name ILIKE ? OR products.description ILIKE ? OR categories.name ILIKE ? OR product_statuses.name ILIKE ? OR product_conditions.name ILIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			).
			Group("products.id") // Group by to avoid duplicates from joins
	}

	// Get total count with search filter applied
	if err := baseQuery.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Build query for fetching products
	// Use a subquery approach to get matching IDs first, then fetch with Preload
	if search != "" && strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		var productIDs []uint
		if err := database.DB.Model(&models.Product{}).
			Joins("LEFT JOIN categories ON products.category_id = categories.id").
			Joins("LEFT JOIN product_statuses ON products.status_id = product_statuses.id").
			Joins("LEFT JOIN product_conditions ON products.product_condition_id = product_conditions.id").
			Where(
				"products.name ILIKE ? OR products.description ILIKE ? OR categories.name ILIKE ? OR product_statuses.name ILIKE ? OR product_conditions.name ILIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			).
			Group("products.id").
			Pluck("products.id", &productIDs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error: " + err.Error(),
			})
			return
		}

		// If no matching IDs, return empty result
		if len(productIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"message": "Products paginated successfully",
				"products": gin.H{
					"total": 0,
					"items": []models.Product{},
				},
			})
			return
		}

		// Fetch products by IDs with Preload
		if err := database.DB.Preload("Status").Preload("ProductCondition").Preload("Category").
			Where("products.id IN ?", productIDs).
			Order("id asc").
			Offset(offset).
			Limit(limit).
			Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error: " + err.Error(),
			})
			return
		}
	} else {
		// No search - fetch all with Preload
		if err := database.DB.Preload("Status").Preload("ProductCondition").Preload("Category").
			Order("id asc").
			Offset(offset).
			Limit(limit).
			Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Database error: " + err.Error(),
			})
			return
		}
	}

	// Get active boosts for all products
	var productIDsForBoost []uint
	for _, p := range products {
		productIDsForBoost = append(productIDsForBoost, p.ID)
	}

	var activeBoosts []models.Boost
	if len(productIDsForBoost) > 0 {
		now := time.Now()
		database.DB.Where("item_type = ? AND item_id IN ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			"product", productIDsForBoost, "active", now, now).Find(&activeBoosts)
	}

	// Create a map of product ID to boost
	boostMap := make(map[uint]*models.Boost)
	for i := range activeBoosts {
		boostMap[activeBoosts[i].ItemID] = &activeBoosts[i]
	}

	// Build response with boost information
	type ProductWithBoost struct {
		models.Product
		ActiveBoost *models.Boost `json:"active_boost,omitempty"`
		IsBoosted   bool          `json:"is_boosted"`
	}

	baseURL := utils.GetBaseURL(c)
	itemsWithBoost := make([]ProductWithBoost, len(products))
	for i, p := range products {
		boost, hasBoost := boostMap[p.ID]
		pCopy := p
		pCopy.Images = utils.FormatImageURLs(p.Images, baseURL)
		itemsWithBoost[i] = ProductWithBoost{
			Product:     pCopy,
			ActiveBoost: boost,
			IsBoosted:   hasBoost,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Products paginated successfully",
		"products": gin.H{
			"total": total,
			"items": itemsWithBoost,
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

	// Get product before deletion for notification
	var product models.Product
	if err := database.DB.Where("id = ? AND seller_id = ?", idInt, userID).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found or access denied"})
		return
	}

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

	// Notify admins about product deletion
	reference := fmt.Sprintf("product:%d", product.ID)
	services.NotifyAdmins(
		"Product Deleted",
		fmt.Sprintf("Product '%s' has been deleted by the seller", product.Name),
		"warning",
		&reference,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
	})
}

func ListingDetails(c *gin.Context) {
	id := c.Param("id")
	itemType := c.Query("type")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	// If type is explicitly provided, try that first
	if itemType == "product" {
		var product models.Product
		if err := database.DB.Preload("Status").Preload("ProductCondition").Preload("Category").Preload("Seller").First(&product, idInt).Error; err == nil {
			baseURL := utils.GetBaseURL(c)
			c.JSON(http.StatusOK, gin.H{
				"message": "Product details retrieved successfully",
				"item": gin.H{
					"id":                product.ID,
					"name":              product.Name,
					"description":       product.Description,
					"price":             product.Price,
					"location":          product.Location,
					"images":            utils.FormatImageURLs(product.Images, baseURL),
					"item_type":         "product",
					"status":            product.Status,
					"product_condition": product.ProductCondition,
					"category":          product.Category,
					"seller":            product.Seller,
					"CreatedAt":         product.CreatedAt,
				},
			})
			return
		}
	} else if itemType == "vehicle" {
		var vehicle models.Vehicle
		if err := database.DB.Preload("VehicleType").Preload("Seller").First(&vehicle, idInt).Error; err == nil {
			baseURL := utils.GetBaseURL(c)
			vehicleName := fmt.Sprintf("%s %s %d", vehicle.VehicleMake, vehicle.VehicleModel, vehicle.Year)
			c.JSON(http.StatusOK, gin.H{
				"message": "Vehicle details retrieved successfully",
				"item": gin.H{
					"id":            vehicle.ID,
					"name":          vehicleName,
					"description":   vehicle.Description,
					"price":         float64(vehicle.Price),
					"location":      vehicle.Location,
					"images":        utils.FormatImageURLs(vehicle.Images, baseURL),
					"item_type":     "vehicle",
					"vehicle_type":  vehicle.VehicleType,
					"vehicle_make":  vehicle.VehicleMake,
					"vehicle_model": vehicle.VehicleModel,
					"year":          vehicle.Year,
					"seller":        vehicle.Seller,
					"CreatedAt":     vehicle.CreatedAt,
				},
			})
			return
		}
	}

	// Fallback/Legacy behavior: Try product first, then vehicle
	if itemType == "" {
		// Try product first
		var product models.Product
		if err := database.DB.Preload("Status").Preload("ProductCondition").Preload("Category").Preload("Seller").First(&product, idInt).Error; err == nil {
			baseURL := utils.GetBaseURL(c)
			c.JSON(http.StatusOK, gin.H{
				"message": "Product details retrieved successfully",
				"item": gin.H{
					"id":                product.ID,
					"name":              product.Name,
					"description":       product.Description,
					"price":             product.Price,
					"location":          product.Location,
					"images":            utils.FormatImageURLs(product.Images, baseURL),
					"item_type":         "product",
					"status":            product.Status,
					"product_condition": product.ProductCondition,
					"category":          product.Category,
					"seller":            product.Seller,
					"CreatedAt":         product.CreatedAt,
				},
			})
			return
		}

		// Try vehicle
		var vehicle models.Vehicle
		if err := database.DB.Preload("VehicleType").Preload("Seller").First(&vehicle, idInt).Error; err == nil {
			baseURL := utils.GetBaseURL(c)
			vehicleName := fmt.Sprintf("%s %s %d", vehicle.VehicleMake, vehicle.VehicleModel, vehicle.Year)
			c.JSON(http.StatusOK, gin.H{
				"message": "Vehicle details retrieved successfully",
				"item": gin.H{
					"id":            vehicle.ID,
					"name":          vehicleName,
					"description":   vehicle.Description,
					"price":         float64(vehicle.Price),
					"location":      vehicle.Location,
					"images":        utils.FormatImageURLs(vehicle.Images, baseURL),
					"item_type":     "vehicle",
					"vehicle_type":  vehicle.VehicleType,
					"vehicle_make":  vehicle.VehicleMake,
					"vehicle_model": vehicle.VehicleModel,
					"year":          vehicle.Year,
					"seller":        vehicle.Seller,
					"CreatedAt":     vehicle.CreatedAt,
				},
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"message": "Listing not found"})
}

func ProductDetails(c *gin.Context) {
	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid product ID"})
		return
	}

	var product models.Product
	if err := database.DB.Preload("Status").Preload("ProductCondition").Preload("Category").Preload("Seller").First(&product, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
		return
	}

	baseURL := utils.GetBaseURL(c)
	product.Images = utils.FormatImageURLs(product.Images, baseURL)

	c.JSON(http.StatusOK, gin.H{
		"message": "Product details retrieved successfully",
		"product": product,
	})
}

func GetMyListings(c *gin.Context) {
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

	// Fetch products
	var products []models.Product
	var productTotal int64
	productQuery := database.DB.Model(&models.Product{}).Where("seller_id = ?", userID)
	if err := productQuery.Count(&productTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}
	if err := productQuery.Preload("Status").Preload("ProductCondition").Preload("Category").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Fetch vehicles
	var vehicles []models.Vehicle
	var vehicleTotal int64
	vehicleQuery := database.DB.Model(&models.Vehicle{}).Where("seller_id = ?", userID)
	if err := vehicleQuery.Count(&vehicleTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}
	if err := vehicleQuery.Preload("VehicleType").Limit(limit).Offset(offset).Find(&vehicles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Get active boosts for products
	var productIDsForBoost []uint
	for _, p := range products {
		productIDsForBoost = append(productIDsForBoost, p.ID)
	}
	var vehicleIDsForBoost []uint
	for _, v := range vehicles {
		vehicleIDsForBoost = append(vehicleIDsForBoost, v.ID)
	}

	var activeBoosts []models.Boost
	now := time.Now()
	if len(productIDsForBoost) > 0 {
		database.DB.Where("item_type = ? AND item_id IN ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			"product", productIDsForBoost, "active", now, now).Find(&activeBoosts)
	}
	if len(vehicleIDsForBoost) > 0 {
		var vehicleBoosts []models.Boost
		database.DB.Where("item_type = ? AND item_id IN ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			"vehicle", vehicleIDsForBoost, "active", now, now).Find(&vehicleBoosts)
		activeBoosts = append(activeBoosts, vehicleBoosts...)
	}

	// Create boost maps
	productBoostMap := make(map[uint]*models.Boost)
	vehicleBoostMap := make(map[uint]*models.Boost)
	for i := range activeBoosts {
		if activeBoosts[i].ItemType == "product" {
			productBoostMap[activeBoosts[i].ItemID] = &activeBoosts[i]
		} else if activeBoosts[i].ItemType == "vehicle" {
			vehicleBoostMap[activeBoosts[i].ItemID] = &activeBoosts[i]
		}
	}

	// Build response with boost information for products
	type ProductWithBoost struct {
		models.Product
		ActiveBoost *models.Boost `json:"active_boost,omitempty"`
		IsBoosted   bool          `json:"is_boosted"`
		ItemType    string        `json:"item_type"`
	}

	// Build response for vehicles (convert to product-like structure)
	type VehicleAsProduct struct {
		ID               uint                     `json:"id"`
		Name             string                   `json:"name"`
		Description      string                   `json:"description"`
		Price            float64                  `json:"price"`
		Location         string                   `json:"location"`
		Images           models.StringArray       `json:"images"`
		Category         *models.Category         `json:"category,omitempty"`
		Status           *models.ProductStatus    `json:"status,omitempty"`
		ProductCondition *models.ProductCondition `json:"product_condition,omitempty"`
		CreatedAt        time.Time                `json:"CreatedAt"`
		UpdatedAt        time.Time                `json:"UpdatedAt"`
		ActiveBoost      *models.Boost            `json:"active_boost,omitempty"`
		IsBoosted        bool                     `json:"is_boosted"`
		ItemType         string                   `json:"item_type"`
		VehicleType      *models.VehicleType      `json:"vehicle_type,omitempty"`
		VehicleMake      string                   `json:"vehicle_make,omitempty"`
		VehicleModel     string                   `json:"vehicle_model,omitempty"`
		Year             uint                     `json:"year,omitempty"`
	}

	itemsWithBoost := make([]interface{}, 0, len(products)+len(vehicles))

	baseURL := utils.GetBaseURL(c)
	// Add products
	for _, p := range products {
		boost, hasBoost := productBoostMap[p.ID]
		pCopy := p
		pCopy.Images = utils.FormatImageURLs(p.Images, baseURL)
		itemsWithBoost = append(itemsWithBoost, ProductWithBoost{
			Product:     pCopy,
			ActiveBoost: boost,
			IsBoosted:   hasBoost,
			ItemType:    "product",
		})
	}

	// Add vehicles (convert to product-like structure)
	for _, v := range vehicles {
		boost, hasBoost := vehicleBoostMap[v.ID]
		vehicleName := fmt.Sprintf("%s %s %d", v.VehicleMake, v.VehicleModel, v.Year)
		itemsWithBoost = append(itemsWithBoost, VehicleAsProduct{
			ID:           v.ID,
			Name:         vehicleName,
			Description:  v.Description,
			Price:        float64(v.Price),
			Location:     v.Location,
			Images:       utils.FormatImageURLs(v.Images, baseURL),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
			ActiveBoost:  boost,
			IsBoosted:    hasBoost,
			ItemType:     "vehicle",
			VehicleType:  &v.VehicleType,
			VehicleMake:  v.VehicleMake,
			VehicleModel: v.VehicleModel,
			Year:         v.Year,
		})
	}

	total := productTotal + vehicleTotal

	c.JSON(http.StatusOK, gin.H{
		"message": "My listings retrieved successfully",
		"products": gin.H{
			"total": total,
			"items": itemsWithBoost,
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
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if err := query.Preload("Status").Preload("ProductCondition").Preload("Category").
		Limit(limit).Offset(offset).Order("id desc").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	baseURL := utils.GetBaseURL(c)
	for i := range products {
		products[i].Images = utils.FormatImageURLs(products[i].Images, baseURL)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "My products retrieved successfully",
		"products": gin.H{
			"total": total,
			"items": products,
		},
	})
}

func GetAllListings(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	search := c.Query("search")
	categoryIDs := c.Query("category_ids")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	conditionIDs := c.Query("condition_ids")
	sort := c.Query("sort")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Fetch products
	var products []models.Product
	var productTotal int64
	productQuery := database.DB.Model(&models.Product{})

	if search != "" {
		searchTerm := "%" + search + "%"
		productQuery = productQuery.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}

	if categoryIDs != "" {
		ids := strings.Split(categoryIDs, ",")
		productQuery = productQuery.Where("category_id IN ?", ids)
	}

	if minPrice != "" {
		price, _ := strconv.ParseFloat(minPrice, 64)
		productQuery = productQuery.Where("price >= ?", price)
	}

	if maxPrice != "" {
		price, _ := strconv.ParseFloat(maxPrice, 64)
		productQuery = productQuery.Where("price <= ?", price)
	}

	if conditionIDs != "" {
		ids := strings.Split(conditionIDs, ",")
		productQuery = productQuery.Where("product_condition_id IN ?", ids)
	}

	// Sorting
	switch sort {
	case "price_asc":
		productQuery = productQuery.Order("price ASC")
	case "price_desc":
		productQuery = productQuery.Order("price DESC")
	case "newest":
		productQuery = productQuery.Order("created_at DESC")
	default:
		productQuery = productQuery.Order("id DESC")
	}

	if err := productQuery.Count(&productTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}
	if err := productQuery.Preload("Status").Preload("ProductCondition").Preload("Category").Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	// Fetch vehicles
	var vehicles []models.Vehicle
	var vehicleTotal int64
	vehicleQuery := database.DB.Model(&models.Vehicle{})

	if search != "" {
		searchTerm := "%" + search + "%"
		vehicleQuery = vehicleQuery.Where("vehicle_make ILIKE ? OR vehicle_model ILIKE ? OR description ILIKE ?", searchTerm, searchTerm, searchTerm)
	}

	// Check if this is a vehicle category
	isVehicleOnly := false
	if categoryIDs != "" {
		ids := strings.Split(categoryIDs, ",")
		var categoryObjects []models.Category
		if err := database.DB.Preload("CategoryGroup").Where("id IN ?", ids).Find(&categoryObjects).Error; err == nil {
			hasVehicleCategory := false
			hasOtherCategory := false
			for _, cat := range categoryObjects {
				if cat.CategoryGroup.Name == "Vehicles" || cat.Name == "Vehicles" {
					hasVehicleCategory = true
				} else {
					hasOtherCategory = true
				}
			}

			if hasVehicleCategory && !hasOtherCategory {
				isVehicleOnly = true
			} else if !hasVehicleCategory {
				// If filtering by non-vehicle categories only, hide vehicles
				vehicleQuery = vehicleQuery.Where("1 = 0")
			}
		}
	}

	if minPrice != "" {
		price, _ := strconv.ParseFloat(minPrice, 64)
		vehicleQuery = vehicleQuery.Where("price >= ?", price)
	}

	if maxPrice != "" {
		price, _ := strconv.ParseFloat(maxPrice, 64)
		vehicleQuery = vehicleQuery.Where("price <= ?", price)
	}

	// Vehicles don't have conditions in the current model
	if conditionIDs != "" {
		vehicleQuery = vehicleQuery.Where("1 = 0")
	}

	// Sorting for vehicles
	switch sort {
	case "price_asc":
		vehicleQuery = vehicleQuery.Order("price ASC")
	case "price_desc":
		vehicleQuery = vehicleQuery.Order("price DESC")
	case "newest":
		vehicleQuery = vehicleQuery.Order("created_at DESC")
	default:
		vehicleQuery = vehicleQuery.Order("id DESC")
	}

	if err := vehicleQuery.Count(&vehicleTotal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}
	if err := vehicleQuery.Preload("VehicleType").Limit(limit).Offset(offset).Find(&vehicles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	// If it's a vehicle only search, we can zero out products (optional but clearer)
	if isVehicleOnly {
		products = []models.Product{}
		productTotal = 0
	}

	// Get active boosts
	var productIDsForBoost []uint
	for _, p := range products {
		productIDsForBoost = append(productIDsForBoost, p.ID)
	}
	var vehicleIDsForBoost []uint
	for _, v := range vehicles {
		vehicleIDsForBoost = append(vehicleIDsForBoost, v.ID)
	}

	var activeBoosts []models.Boost
	now := time.Now()
	if len(productIDsForBoost) > 0 {
		database.DB.Where("item_type = ? AND item_id IN ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			"product", productIDsForBoost, "active", now, now).Find(&activeBoosts)
	}
	if len(vehicleIDsForBoost) > 0 {
		var vehicleBoosts []models.Boost
		database.DB.Where("item_type = ? AND item_id IN ? AND status = ? AND start_date <= ? AND (end_date IS NULL OR end_date >= ?)",
			"vehicle", vehicleIDsForBoost, "active", now, now).Find(&vehicleBoosts)
		activeBoosts = append(activeBoosts, vehicleBoosts...)
	}

	// Create boost maps
	productBoostMap := make(map[uint]*models.Boost)
	vehicleBoostMap := make(map[uint]*models.Boost)
	for i := range activeBoosts {
		if activeBoosts[i].ItemType == "product" {
			productBoostMap[activeBoosts[i].ItemID] = &activeBoosts[i]
		} else if activeBoosts[i].ItemType == "vehicle" {
			vehicleBoostMap[activeBoosts[i].ItemID] = &activeBoosts[i]
		}
	}

	// Build response items
	type ProductWithBoost struct {
		models.Product
		ActiveBoost *models.Boost `json:"active_boost,omitempty"`
		IsBoosted   bool          `json:"is_boosted"`
		ItemType    string        `json:"item_type"`
	}

	type VehicleAsProduct struct {
		ID               uint                     `json:"id"`
		Name             string                   `json:"name"`
		Description      string                   `json:"description"`
		Price            float64                  `json:"price"`
		Location         string                   `json:"location"`
		Images           models.StringArray       `json:"images"`
		Category         *models.Category         `json:"category,omitempty"`
		Status           *models.ProductStatus    `json:"status,omitempty"`
		ProductCondition *models.ProductCondition `json:"product_condition,omitempty"`
		CreatedAt        time.Time                `json:"CreatedAt"`
		UpdatedAt        time.Time                `json:"UpdatedAt"`
		ActiveBoost      *models.Boost            `json:"active_boost,omitempty"`
		IsBoosted        bool                     `json:"is_boosted"`
		ItemType         string                   `json:"item_type"`
		VehicleType      *models.VehicleType      `json:"vehicle_type,omitempty"`
		VehicleMake      string                   `json:"vehicle_make,omitempty"`
		VehicleModel     string                   `json:"vehicle_model,omitempty"`
		Year             uint                     `json:"year,omitempty"`
	}

	items := make([]interface{}, 0, len(products)+len(vehicles))

	baseURL := utils.GetBaseURL(c)
	for _, p := range products {
		boost, hasBoost := productBoostMap[p.ID]
		// Create a copy of the product to avoid mutating the original if it's cached/shared
		pCopy := p
		pCopy.Images = utils.FormatImageURLs(p.Images, baseURL)
		items = append(items, ProductWithBoost{
			Product:     pCopy,
			ActiveBoost: boost,
			IsBoosted:   hasBoost,
			ItemType:    "product",
		})
	}

	for _, v := range vehicles {
		boost, hasBoost := vehicleBoostMap[v.ID]
		vehicleName := fmt.Sprintf("%s %s %d", v.VehicleMake, v.VehicleModel, v.Year)
		items = append(items, VehicleAsProduct{
			ID:           v.ID,
			Name:         vehicleName,
			Description:  v.Description,
			Price:        float64(v.Price),
			Location:     v.Location,
			Images:       utils.FormatImageURLs(v.Images, baseURL),
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
			ActiveBoost:  boost,
			IsBoosted:    hasBoost,
			ItemType:     "vehicle",
			VehicleType:  &v.VehicleType,
			VehicleMake:  v.VehicleMake,
			VehicleModel: v.VehicleModel,
			Year:         v.Year,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Listings retrieved successfully",
		"listings": gin.H{
			"total": productTotal + vehicleTotal,
			"items": items,
		},
	})
}

func RevealContact(c *gin.Context) {
	id := c.Param("id")
	itemType := c.Query("type")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	userID := c.MustGet("userID").(uint)

	// Basic rate limiting: Check if user has viewed too many contacts in the last hour
	var count int64
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	database.DB.Model(&models.ContactViewLog{}).Where("viewer_id = ? AND created_at > ?", userID, oneHourAgo).Count(&count)

	if count > 5 {
		c.JSON(http.StatusTooManyRequests, gin.H{"message": "You've reached the limit for viewing contact info. Please try again later."})
		return
	}

	var sellerID uint
	var phone, email string

	if itemType == "vehicle" {
		var vehicle models.Vehicle
		if err := database.DB.Preload("Seller").First(&vehicle, idInt).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Vehicle not found"})
			return
		}
		sellerID = vehicle.SellerID
		phone = vehicle.Seller.Phone
		email = vehicle.Seller.Email
	} else {
		var product models.Product
		if err := database.DB.Preload("Seller").First(&product, idInt).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
			return
		}
		sellerID = product.SellerID
		phone = product.Seller.Phone
		email = product.Seller.Email
	}

	// Log the event
	log := models.ContactViewLog{
		ViewerID:  userID,
		SellerID:  sellerID,
		ProductID: uint(idInt),
	}
	database.DB.Create(&log)

	c.JSON(http.StatusOK, gin.H{
		"phone": phone,
		"email": email,
	})
}
