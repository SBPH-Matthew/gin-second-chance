package handlers

import (
	"net/http"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
	"github.com/gin-gonic/gin"
)

func CreateReview(c *gin.Context) {
	type ReviewRequest struct {
		TargetUserID uint   `json:"target_user_id" binding:"required"`
		ProductID    uint   `json:"product_id"`
		Rating       int    `json:"rating" binding:"required,min=1,max=5"`
		Comment      string `json:"comment"`
	}

	var body ReviewRequest
	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	reviewerID := c.MustGet("userID").(uint)

	review := models.Review{
		ReviewerID:   reviewerID,
		TargetUserID: body.TargetUserID,
		ProductID:    body.ProductID,
		Rating:       body.Rating,
		Comment:      body.Comment,
	}

	if err := database.DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create review"})
		return
	}

	// Update target user's aggregated rating
	var result struct {
		AvgRating float64
		Total     int64
	}
	database.DB.Model(&models.Review{}).
		Select("AVG(rating) as avg_rating, COUNT(*) as total").
		Where("target_user_id = ?", body.TargetUserID).
		Scan(&result)

	database.DB.Model(&models.User{}).Where("id = ?", body.TargetUserID).Updates(map[string]interface{}{
		"rating":        result.AvgRating,
		"total_reviews": result.Total,
	})

	c.JSON(http.StatusOK, review)
}

func GetUserReviews(c *gin.Context) {
	targetUserID := c.Param("id")

	var reviews []models.Review
	if err := database.DB.Preload("Reviewer").
		Where("target_user_id = ?", targetUserID).
		Order("created_at desc").
		Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch reviews"})
		return
	}

	baseURL := utils.GetBaseURL(c)
	for i := range reviews {
		reviews[i].Reviewer.ProfilePicture = utils.FormatImageURL(reviews[i].Reviewer.ProfilePicture, baseURL)
	}

	c.JSON(http.StatusOK, reviews)
}
