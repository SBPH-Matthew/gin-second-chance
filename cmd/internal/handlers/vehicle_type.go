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

func CreateVehicleType(c *gin.Context) {
	body := requests.CreateVehicleTypeRequest{}
	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	vehicleType := models.VehicleType{
		Name: body.Name,
	}

	if err := database.DB.Create(&vehicleType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vehicle type created successfully",
	})
}

func GetAllVehicleTypes(c *gin.Context) {
	var vehicleTypes []models.VehicleType
	var total int64

	if err := database.DB.Model(&models.VehicleType{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	if err := database.DB.Order("id asc").Find(&vehicleTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle types retrieved successfully",
		"vehicleTypes": gin.H{
			"total": total,
			"items": vehicleTypes,
		},
	})
}

func UpdateVehicleType(c *gin.Context) {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	body := requests.CreateVehicleTypeRequest{}
	if err := utils.ValidateBodyJSON(c, &body); err != nil {
		return
	}

	var vehicleType models.VehicleType
	if err := database.DB.First(&vehicleType, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Vehicle type not found"})
		return
	}

	vehicleType.Name = body.Name

	if err := database.DB.Save(&vehicleType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vehicle type updated successfully",
	})
}

func DeleteVehicleType(c *gin.Context) {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid ID"})
		return
	}

	var vehicleType models.VehicleType
	if err := database.DB.First(&vehicleType, idInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Vehicle type not found"})
		return
	}

	if err := database.DB.Delete(&vehicleType).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vehicle type deleted successfully"})
}
