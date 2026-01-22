package handlers

import (
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
	"golang.org/x/crypto/bcrypt"
)

func GetUsers(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"users": []string{"John Doe", "Jane Doe"},
	})
}

func GetPaginateUser(c *gin.Context) {
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

	var users []models.User
	var total int64

	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	if err := database.DB.Preload("Role").Order("id asc").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User retrieved successfully",
		"users": gin.H{
			"total": total,
			"items": users,
		},
	})
}

func CreateUser(c *gin.Context) {
	// Parse multipart form for file upload
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse form data"})
		return
	}

	type CreateUserRequest struct {
		FirstName       string `form:"first_name" validate:"required"`
		LastName        string `form:"last_name" validate:"required"`
		Email           string `form:"email" validate:"required,email"`
		Role            string `form:"role" validate:"required"`
		Password        string `form:"password" validate:"required,min=8,max=100"`
		ConfirmPassword string `form:"confirm_password" validate:"required,min=8,max=100,eqfield=Password"`
		// Address fields
		Country        string `form:"country"`
		StateProvince  string `form:"state_province"`
		StreetAddress1 string `form:"street_address_1"`
		StreetAddress2 string `form:"street_address_2"`
		ZipPostalCode  string `form:"zip_postal_code"`
	}

	var body CreateUserRequest

	if err := utils.ValidateBodyFormData(c, &body); err != nil {
		return
	}

	var existingUser models.User
	result := database.DB.Where("email = ?", body.Email).Find(&existingUser)

	// If RowsAffected is greater than 0, it means the user exists
	if result.RowsAffected > 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": body.Email,
			"errors":  gin.H{"email": "This email is already taken"},
		})
		return
	}

	roleID, err := strconv.Atoi(body.Role)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Handle profile picture upload
	var profilePicturePath string
	multipartForm, err := c.MultipartForm()
	if err == nil && multipartForm != nil {
		formFiles := multipartForm.File["profile_picture"]
		if len(formFiles) > 0 {
			file := formFiles[0]
			src, err := file.Open()
			if err == nil {
				defer src.Close()

				// Generate unique filename
				ext := filepath.Ext(file.Filename)
				// Remove duplicate extensions
				baseName := strings.TrimSuffix(file.Filename, ext)
				baseName = strings.TrimSuffix(baseName, ext)
				filename := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)

				// Create uploads directory if it doesn't exist
				uploadDir := getUploadDir()
				profileDir := filepath.Join(uploadDir, "profiles")
				if err := os.MkdirAll(profileDir, os.ModePerm); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message": "Failed to create upload directory: " + err.Error(),
					})
					return
				}

				// Save file
				dst, err := os.Create(filepath.Join(profileDir, filename))
				if err == nil {
					defer dst.Close()
					io.Copy(dst, src)
					profilePicturePath = fmt.Sprintf("/uploads/profiles/%s", filename)
				}
			}
		}
	}

	user := models.User{
		FirstName:      body.FirstName,
		LastName:       body.LastName,
		Email:          body.Email,
		RoleID:         uint(roleID),
		Password:       string(passwordHash),
		ProfilePicture: profilePicturePath,
		Country:        body.Country,
		StateProvince:  body.StateProvince,
		StreetAddress1: body.StreetAddress1,
		StreetAddress2: body.StreetAddress2,
		ZipPostalCode:  body.ZipPostalCode,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	// Create notification for admins about new user
	reference := fmt.Sprintf("user:%d", user.ID)
	services.NotifyAdmins(
		"New User Created",
		fmt.Sprintf("A new user '%s %s' (%s) has been created", user.FirstName, user.LastName, user.Email),
		"info",
		&reference,
	)

	// Notify the new user
	services.NotifyUser(
		user.ID,
		"Account Created",
		fmt.Sprintf("Your account has been created successfully. Welcome, %s!", user.FirstName),
		"success",
		&reference,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
	})
}

func Register(c *gin.Context) {
	var body requests.RegisterRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	// 1. Manual Uniqueness Check
	var existingUser models.User
	if err := database.DB.Where("email = ?", body.Email).First(&existingUser).Error; err == nil {
		// If err is nil, it means a user was found
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{
			"message": "This email is already taken",
			"errors":  gin.H{"email": "This email is already taken"},
		})
		return
	}

	user := models.User{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
	}

	// 🔒 Hash using model method
	if err := user.HashPassword(body.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to hash password",
		})
		return
	}

	// Save to DB
	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Create notification for admins about new registration
	reference := fmt.Sprintf("user:%d", user.ID)
	services.NotifyAdmins(
		"New User Registration",
		fmt.Sprintf("A new user '%s %s' (%s) has registered", user.FirstName, user.LastName, user.Email),
		"info",
		&reference,
	)

	// Notify the new user
	services.NotifyUser(
		user.ID,
		"Registration Successful",
		fmt.Sprintf("Welcome to Second Chance! Your account has been created successfully, %s!", user.FirstName),
		"success",
		&reference,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Registration successful",
	})
}

func Login(c *gin.Context) {
	var body requests.LoginRequest

	if err := c.ShouldBindBodyWithJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	var user models.User
	err := database.DB.Where("email = ?", body.Email).First(&user).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid email or password credentials",
			"errors":  gin.H{"email": "Invalid email or password credentials"}})
		return
	}

	if !user.CheckPassword(body.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid email or password",
			"errors":  gin.H{"email": "Invalid email or password"}})
		return
	}

	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to generate token",
			"errors":  gin.H{"email": "Failed to generate token"}})
		return
	}

	// --- SET THE HTTP-ONLY COOKIE ---
	c.SetCookie(
		"token",     // Name of the cookie
		token,       // The JWT string
		3600*24,     // MaxAge in seconds (e.g., 24 hours)
		"/",         // Path (available to all routes)
		"localhost", // Domain (change to your actual domain in production)
		false,       // Secure: Set to true if using HTTPS (essential for production!)
		true,        // HttpOnly: TRUE (prevents JavaScript access)
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
	})
}

func UpdateUser(c *gin.Context) {
	// Parse multipart form for file upload
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		c.JSON(http.StatusBadRequest, gin.H{"message": "Failed to parse form data"})
		return
	}

	type UpdateUserRequest struct {
		FirstName       string `form:"first_name" binding:"required"`
		LastName        string `form:"last_name" binding:"required"`
		Email           string `form:"email" binding:"required,email"`
		Role            string `form:"role" binding:"required"`
		// Address fields
		Country        string `form:"country"`
		StateProvince  string `form:"state_province"`
		StreetAddress1 string `form:"street_address_1"`
		StreetAddress2 string `form:"street_address_2"`
		ZipPostalCode  string `form:"zip_postal_code"`
		// Existing profile picture path (if not uploading new one)
		ExistingProfilePicture string `form:"existing_profile_picture"`
	}

	var body UpdateUserRequest

	if err := c.ShouldBind(&body); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request data",
			"errors":  err.Error(),
		})
		return
	}

	id := c.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid user id",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, idInt).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	var count int64
	database.DB.Model(&models.User{}).
		Where("email = ? AND id <> ?", body.Email, idInt).
		Count(&count)

	if count > 0 {
		c.AbortWithStatusJSON(http.StatusConflict, gin.H{ // 409 Conflict is better than 500
			"message": "Email already exists",
			"errors":  gin.H{"email": "This email is already registered to another account"},
		})
		return
	}

	roleID, err := strconv.Atoi(body.Role)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid role id",
		})
		return
	}

	// Handle profile picture upload
	profilePicturePath := body.ExistingProfilePicture
	multipartForm, err := c.MultipartForm()
	if err == nil && multipartForm != nil {
		formFiles := multipartForm.File["profile_picture"]
		if len(formFiles) > 0 {
			// Delete old profile picture if exists
			if user.ProfilePicture != "" {
				relativePath := strings.TrimPrefix(user.ProfilePicture, "/uploads/")
				fullPath := filepath.Join(getUploadDir(), relativePath)
				if err := os.Remove(fullPath); err != nil {
					// Log error but don't fail the update
					fmt.Printf("Failed to delete profile picture file %s: %v\n", fullPath, err)
				}
			}

			file := formFiles[0]
			src, err := file.Open()
			if err == nil {
				defer src.Close()

				// Generate unique filename
				ext := filepath.Ext(file.Filename)
				// Remove duplicate extensions
				baseName := strings.TrimSuffix(file.Filename, ext)
				baseName = strings.TrimSuffix(baseName, ext)
				filename := fmt.Sprintf("%s_%d%s", baseName, time.Now().UnixNano(), ext)

				// Create uploads directory if it doesn't exist
				uploadDir := getUploadDir()
				profileDir := filepath.Join(uploadDir, "profiles")
				if err := os.MkdirAll(profileDir, os.ModePerm); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{
						"message": "Failed to create upload directory: " + err.Error(),
					})
					return
				}

				// Save file
				dst, err := os.Create(filepath.Join(profileDir, filename))
				if err == nil {
					defer dst.Close()
					io.Copy(dst, src)
					profilePicturePath = fmt.Sprintf("/uploads/profiles/%s", filename)
				}
			}
		}
	}

	updateUser := models.User{
		ID:             uint(idInt),
		FirstName:      body.FirstName,
		LastName:       body.LastName,
		Email:          body.Email,
		RoleID:         uint(roleID),
		ProfilePicture: profilePicturePath,
		Country:        body.Country,
		StateProvince:  body.StateProvince,
		StreetAddress1: body.StreetAddress1,
		StreetAddress2: body.StreetAddress2,
		ZipPostalCode:  body.ZipPostalCode,
	}

	if err := database.DB.Model(&updateUser).Updates(updateUser).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User details updated!",
	})
}

func ChangeUserPassword(c *gin.Context) {
	var body requests.UpdateUserPasswordRequest

	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Invalid ID",
		})
		return
	}

	var user models.User
	if err := database.DB.First(&user, idInt).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	if err := user.CheckPassword(body.OldPassword); err == false {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Old password is invalid",
			"errors": gin.H{
				"old_password": "Old password is invalid",
			},
		})
		return
	}

	if body.NewPassword != body.ConfirmPassword {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "Password confirmation does not match",
			"errors": gin.H{
				"confirm_password": "Password confirmation does not match",
			},
		})
		return
	}

	if err := user.HashPassword(body.NewPassword); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if err := database.DB.Model(&user).
		Update("password", user.Password).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}

func DeleteUser(c *gin.Context) {
	idInt, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": "No id is provided",
		})
		return
	}

	// Get user before deletion for notification
	var user models.User
	if err := database.DB.First(&user, idInt).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "Database error: " + err.Error(),
		})
		return
	}

	// Notify admins about user deletion
	reference := fmt.Sprintf("user:%d", user.ID)
	services.NotifyAdmins(
		"User Deleted",
		fmt.Sprintf("User '%s %s' (%s) has been deleted", user.FirstName, user.LastName, user.Email),
		"warning",
		&reference,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})

}
