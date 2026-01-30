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

	itemsWithBoost := make([]ProductWithBoost, len(products))
	for i, p := range products {
		boost, hasBoost := boostMap[p.ID]
		itemsWithBoost[i] = ProductWithBoost{
			Product:     p,
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

	// Add products
	for _, p := range products {
		boost, hasBoost := productBoostMap[p.ID]
		itemsWithBoost = append(itemsWithBoost, ProductWithBoost{
			Product:     p,
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
			Images:       v.Images,
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
