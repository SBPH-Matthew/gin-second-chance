package models

import "gorm.io/gorm"

type Vehicle struct {
	gorm.Model
	ID            uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	VehicleTypeID *uint       `json:"vehicle_type_id"`
	Location      string      `json:"location"`
	Year          uint        `json:"year"`
	VehicleMake   string      `json:"vehicle_make"`
	VehicleModel  string      `json:"vehicle_model"`
	Price         uint        `json:"price"`
	Description   string      `json:"description"`
	Images        StringArray `gorm:"type:json" json:"images"`
	SellerID      uint        `json:"seller_id"`

	VehicleType VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicle_type"`
	Seller      User        `gorm:"foreignKey:SellerID" json:"seller"`
}
