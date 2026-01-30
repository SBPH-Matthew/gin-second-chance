package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/gin-gonic/gin"
)

func VerifyIdentity(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	// Parse multipart form for file upload
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse form data"})
		return
	}

	// Get the uploaded file
	multipartForm, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse multipart form"})
		return
	}

	formFiles := multipartForm.File["id_file"]
	if len(formFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "ID document file is required"})
		return
	}

	file := formFiles[0]

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := []string{".jpg", ".jpeg", ".png", ".pdf"}
	isValidExt := false
	for _, validExt := range validExts {
		if ext == validExt {
			isValidExt = true
			break
		}
	}

	if !isValidExt {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid file type. Please upload a JPEG, PNG, or PDF file"})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "File size must be less than 5MB"})
		return
	}

	// Get existing user to check for old ID document
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
		return
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	// Generate unique filename
	baseName := strings.TrimSuffix(file.Filename, ext)
	baseName = strings.TrimSuffix(baseName, ext) // Remove duplicate extensions
	filename := fmt.Sprintf("id_%d_%d%s", userID, time.Now().UnixNano(), ext)

	// Create uploads directory if it doesn't exist
	uploadDir := getUploadDir()
	identityDir := filepath.Join(uploadDir, "identity")
	if err := os.MkdirAll(identityDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create upload directory: " + err.Error(),
		})
		return
	}

	// Delete old ID document if exists
	if user.IDDocument != "" {
		relativePath := strings.TrimPrefix(user.IDDocument, "/uploads/")
		fullPath := filepath.Join(uploadDir, relativePath)
		if err := os.Remove(fullPath); err != nil {
			// Log error but don't fail the update
			fmt.Printf("Failed to delete old ID document file %s: %v\n", fullPath, err)
		}
	}

	// Save file
	dst, err := os.Create(filepath.Join(identityDir, filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create file: " + err.Error()})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save file: " + err.Error()})
		return
	}

	// Save the file path to database
	// Note: identity_verified remains false until admin reviews and approves
	idDocumentPath := fmt.Sprintf("/uploads/identity/%s", filename)
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("id_document", idDocumentPath).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to save ID document path"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ID document uploaded successfully. Your verification is under review.",
	})
}

func UpdatePhone(c *gin.Context) {
	type PhoneRequest struct {
		Phone string `json:"phone" binding:"required"`
	}

	var body PhoneRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid phone number"})
		return
	}

	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User not found"})
		return
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("phone", body.Phone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update phone"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Phone number updated"})
}