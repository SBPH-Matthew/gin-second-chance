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
	"github.com/SBPH-Matthew/second-chance/cmd/internal/services"
	"github.com/gin-gonic/gin"
)

func CreateVehicle(c *gin.Context) {
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
	vehicleMake := c.PostForm("vehicleMake")
	vehicleModel := c.PostForm("vehicleModel")
	yearStr := c.PostForm("year")
	priceStr := c.PostForm("price")
	description := c.PostForm("description")
	location := c.PostForm("location")
	vehicleTypeStr := c.PostForm("vehicleType")

	// Validate required fields
	if vehicleMake == "" || vehicleModel == "" || yearStr == "" || priceStr == "" || vehicleTypeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing required fields"})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid year"})
		return
	}

	price, err := strconv.Atoi(priceStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid price"})
		return
	}

	vehicleTypeID, err := strconv.Atoi(vehicleTypeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid vehicle type ID"})
		return
	}

	// Handle file uploads
	var imagePaths models.StringArray
	multipartForm, err := c.MultipartForm()
	if err == nil && multipartForm != nil {
		formFiles := multipartForm.File["images"]

		// Create uploads directory if it doesn't exist
		uploadDir := filepath.Join(getUploadDir(), "vehicles")
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
			imagePaths = append(imagePaths, fmt.Sprintf("/uploads/vehicles/%s", filename))
		}
	}

	vehicleTypeIDUint := uint(vehicleTypeID)
	vehicle := models.Vehicle{
		VehicleMake:  vehicleMake,
		VehicleModel: vehicleModel,
		Year:         uint(year),
		Price:        uint(price),
		Description:  description,
		Location:     location,
		Images:       imagePaths,
		VehicleTypeID: &vehicleTypeIDUint,
		SellerID:     user.ID,
	}

	if err := database.DB.Create(&vehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Create notification for admins about new vehicle
	reference := fmt.Sprintf("vehicle:%d", vehicle.ID)
	services.NotifyAdmins(
		"New Vehicle Created",
		fmt.Sprintf("A new vehicle '%s %s' has been created by %s %s", vehicle.VehicleMake, vehicle.VehicleModel, user.FirstName, user.LastName),
		"info",
		&reference,
	)

	// Notify the seller
	services.NotifyUser(
		user.ID,
		"Vehicle Created Successfully",
		fmt.Sprintf("Your vehicle '%s %s' has been created successfully", vehicle.VehicleMake, vehicle.VehicleModel),
		"success",
		&reference,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle created successfully",
	})
}

func UpdateVehicle(c *gin.Context) {
	id := c.Param("id")
	vehicleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid vehicle ID"})
		return
	}

	var existingVehicle models.Vehicle
	if err := database.DB.Preload("VehicleType").Preload("Seller").First(&existingVehicle, vehicleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Vehicle not found"})
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse multipart form: " + err.Error()})
		return
	}

	// Update fields if provided
	if vehicleMake := c.PostForm("vehicleMake"); vehicleMake != "" {
		existingVehicle.VehicleMake = vehicleMake
	}
	if vehicleModel := c.PostForm("vehicleModel"); vehicleModel != "" {
		existingVehicle.VehicleModel = vehicleModel
	}
	if yearStr := c.PostForm("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			existingVehicle.Year = uint(year)
		}
	}
	if priceStr := c.PostForm("price"); priceStr != "" {
		if price, err := strconv.Atoi(priceStr); err == nil {
			existingVehicle.Price = uint(price)
		}
	}
	if description := c.PostForm("description"); description != "" {
		existingVehicle.Description = description
	}
	if location := c.PostForm("location"); location != "" {
		existingVehicle.Location = location
	}
	if vehicleTypeStr := c.PostForm("vehicleType"); vehicleTypeStr != "" {
		if vehicleTypeID, err := strconv.Atoi(vehicleTypeStr); err == nil {
			vehicleTypeIDUint := uint(vehicleTypeID)
			existingVehicle.VehicleTypeID = &vehicleTypeIDUint
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
		imagesToKeep = existingVehicle.Images
	}

	// Delete files that are no longer in the images to keep
	for _, oldImagePath := range existingVehicle.Images {
		shouldKeep := false
		for _, keepPath := range imagesToKeep {
			if oldImagePath == keepPath {
				shouldKeep = true
				break
			}
		}
		if !shouldKeep {
			// Remove the file from storage
			// oldImagePath is like "/uploads/vehicles/filename.jpg"
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
			uploadDir := filepath.Join(getUploadDir(), "vehicles")
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
				newImagePaths = append(newImagePaths, fmt.Sprintf("/uploads/vehicles/%s", filename))
			}
		}
	}

	// Update images: keep existing ones that weren't deleted + add new ones
	// If existingImagesJSON was provided (even if empty), use it; otherwise keep existing
	if existingImagesJSON != "" {
		// existingImages field was provided, use it (could be empty array to delete all)
		existingVehicle.Images = append(imagesToKeep, newImagePaths...)
	} else if len(newImagePaths) > 0 {
		// No existingImages field but new images added, append to existing
		existingVehicle.Images = append(existingVehicle.Images, newImagePaths...)
	}
	// If neither existingImages nor new images, keep existing images unchanged

	if err := database.DB.Save(&existingVehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle updated successfully",
	})
}

func VehiclePaginate(c *gin.Context) {
	page := c.Query("page")
	limit := c.Query("limit")
	search := c.Query("search")

	if page == "" || limit == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Missing page or limit query parameter"})
		return
	}

	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid page parameter"})
		return
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid limit parameter"})
		return
	}

	var vehicles []models.Vehicle
	var total int64

	offset := (pageInt - 1) * limitInt

	// Build base query
	query := database.DB.Model(&models.Vehicle{})

	// Apply search filter if provided
	if search != "" && strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		query = query.Joins("LEFT JOIN vehicle_types ON vehicles.vehicle_type_id = vehicle_types.id").
			Where(
				"vehicles.vehicle_make ILIKE ? OR vehicles.vehicle_model ILIKE ? OR vehicles.location ILIKE ? OR vehicles.description ILIKE ? OR CAST(vehicles.year AS TEXT) ILIKE ? OR CAST(vehicles.price AS TEXT) ILIKE ? OR vehicle_types.name ILIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			)
	}

	// Count total with search filter applied
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Build query for fetching vehicles with search filter
	fetchQuery := database.DB.Preload("VehicleType").Preload("Seller")

	// Apply search filter to fetch query if provided
	if search != "" && strings.TrimSpace(search) != "" {
		searchTerm := "%" + strings.TrimSpace(search) + "%"
		fetchQuery = fetchQuery.Joins("LEFT JOIN vehicle_types ON vehicles.vehicle_type_id = vehicle_types.id").
			Where(
				"vehicles.vehicle_make ILIKE ? OR vehicles.vehicle_model ILIKE ? OR vehicles.location ILIKE ? OR vehicles.description ILIKE ? OR CAST(vehicles.year AS TEXT) ILIKE ? OR CAST(vehicles.price AS TEXT) ILIKE ? OR vehicle_types.name ILIKE ?",
				searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm,
			)
	}

	// Fetch vehicles with pagination and preload relationships
	if err := fetchQuery.
		Offset(offset).
		Limit(limitInt).
		Order("id desc").
		Find(&vehicles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicles retrieved successfully",
		"vehicles": gin.H{
			"total": total,
			"items": vehicles,
		},
	})
}

func DeleteVehicle(c *gin.Context) {
	id := c.Param("id")
	vehicleID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid vehicle ID"})
		return
	}

	var vehicle models.Vehicle
	if err := database.DB.First(&vehicle, vehicleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Vehicle not found"})
		return
	}

	// Delete associated image files
	for _, imagePath := range vehicle.Images {
		fullPath := fmt.Sprintf(".%s", imagePath)
		if err := os.Remove(fullPath); err != nil {
			fmt.Printf("Failed to delete image file %s: %v\n", fullPath, err)
		}
	}

	if err := database.DB.Delete(&vehicle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// Notify admins about vehicle deletion
	reference := fmt.Sprintf("vehicle:%d", vehicle.ID)
	services.NotifyAdmins(
		"Vehicle Deleted",
		fmt.Sprintf("Vehicle '%s %s' has been deleted", vehicle.VehicleMake, vehicle.VehicleModel),
		"warning",
		&reference,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle deleted successfully",
	})
}
