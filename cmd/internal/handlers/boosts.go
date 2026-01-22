package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/requests"
	"github.com/gin-gonic/gin"
)

// Boost pricing (in PHP)
var boostPricing = map[string]map[int]float64{
	"premium": {
		24:  50.0,  // 24 hours
		72:  120.0, // 3 days
		168: 250.0, // 7 days
	},
	"featured": {
		24:  75.0,
		72:  180.0,
		168: 375.0,
	},
	"top": {
		24:  100.0,
		72:  240.0,
		168: 500.0,
	},
}

func CreateBoost(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req requests.CreateBoostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	// Validate that the item exists and belongs to the user
	var itemExists bool
	if req.ItemType == "product" {
		var product models.Product
		if err := database.DB.Where("id = ? AND seller_id = ?", req.ItemID, userID).First(&product).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Product not found or access denied"})
			return
		}
		itemExists = true
	} else if req.ItemType == "vehicle" {
		var vehicle models.Vehicle
		if err := database.DB.Where("id = ? AND seller_id = ?", req.ItemID, userID).First(&vehicle).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Vehicle not found or access denied"})
			return
		}
		itemExists = true
	}

	if !itemExists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid item type"})
		return
	}

	// Calculate cost
	cost, exists := boostPricing[req.BoostType][req.DurationHours]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid boost type or duration"})
		return
	}

	// Create and immediately activate boost (no payment required for CMS)
	now := time.Now()
	endDate := now.Add(time.Duration(req.DurationHours) * time.Hour)

	boost := models.Boost{
		ItemType:      req.ItemType,
		ItemID:        req.ItemID,
		SellerID:      userID,
		BoostType:     req.BoostType,
		DurationHours: req.DurationHours,
		Cost:          cost, // Store cost for future reference
		StartDate:     &now,
		EndDate:       &endDate,
		Status:        "active",
	}

	if err := database.DB.Create(&boost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create boost"})
		return
	}

	c.JSON(http.StatusCreated, boost)
}

func GetUserBoosts(c *gin.Context) {
	userID := c.GetUint("user_id")

	var boosts []models.Boost
	if err := database.DB.Preload("Payment").Where("seller_id = ?", userID).Order("created_at DESC").Find(&boosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch boosts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"boosts": boosts})
}

func GetBoost(c *gin.Context) {
	userID := c.GetUint("user_id")
	boostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid boost ID"})
		return
	}

	var boost models.Boost
	if err := database.DB.Preload("Payment").Where("id = ? AND seller_id = ?", uint(boostID), userID).First(&boost).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Boost not found"})
		return
	}

	c.JSON(http.StatusOK, boost)
}

func CancelBoost(c *gin.Context) {
	userID := c.GetUint("user_id")
	boostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid boost ID"})
		return
	}

	var boost models.Boost
	if err := database.DB.Where("id = ? AND seller_id = ?", uint(boostID), userID).First(&boost).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Boost not found"})
		return
	}

	if boost.Status != "pending" && boost.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Cannot cancel this boost"})
		return
	}

	boost.Status = "cancelled"
	if err := database.DB.Save(&boost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to cancel boost"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Boost cancelled successfully"})
}

func ProcessBoostPayment(c *gin.Context) {
	userID := c.GetUint("user_id")
	boostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid boost ID"})
		return
	}

	var req requests.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid payment data"})
		return
	}

	var boost models.Boost
	if err := database.DB.Where("id = ? AND seller_id = ?", uint(boostID), userID).First(&boost).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Boost not found"})
		return
	}

	if boost.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Boost is not in pending status"})
		return
	}

	// For now, simulate payment processing
	// In production, integrate with actual payment provider
	payment := models.Payment{
		BoostID:       boost.ID,
		Amount:        boost.Cost,
		Currency:      "PHP",
		PaymentMethod: req.PaymentMethod,
		Status:        "completed", // Simulate successful payment
		TransactionID: "sim_" + strconv.FormatUint(uint64(boost.ID), 10),
	}

	if err := database.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process payment"})
		return
	}

	// Activate boost
	now := time.Now()
	boost.Status = "active"
	boost.StartDate = &now
	endDate := now.Add(time.Duration(boost.DurationHours) * time.Hour)
	boost.EndDate = &endDate

	if err := database.DB.Save(&boost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to activate boost"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"boost": boost, "payment": payment})
}

func UpdateBoost(c *gin.Context) {
	userID := c.GetUint("user_id")
	boostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid boost ID"})
		return
	}

	var req requests.UpdateBoostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	// Find the boost and verify ownership
	var boost models.Boost
	if err := database.DB.Where("id = ? AND seller_id = ?", uint(boostID), userID).First(&boost).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Boost not found or access denied"})
		return
	}

	// Only allow updating active boosts
	if boost.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Can only update active boosts"})
		return
	}

	// Validate boost type and duration
	cost, exists := boostPricing[req.BoostType][req.DurationHours]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid boost type or duration"})
		return
	}

	// Update the boost
	now := time.Now()
	boost.BoostType = req.BoostType
	boost.DurationHours = req.DurationHours
	boost.Cost = cost
	boost.StartDate = &now
	endDate := now.Add(time.Duration(req.DurationHours) * time.Hour)
	boost.EndDate = &endDate
	// Status remains "active"

	if err := database.DB.Save(&boost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update boost"})
		return
	}

	c.JSON(http.StatusOK, boost)
}

func GetBoostPricing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"pricing": boostPricing})
}
