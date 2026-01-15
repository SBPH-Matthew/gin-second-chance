package models

import "gorm.io/gorm"

type Vehicle struct {
	gorm.Model
	ID          uint        `gorm:"primaryKey;autoIncrement"`
	VehicleTypeID *uint
	Location      string
	Year         uint
	VehicleMake  string
	VehicleModel string
	Price        uint
	Description  string
	Images       StringArray `gorm:"type:json"`
	SellerID     uint

	VehicleType  VehicleType `gorm:"foreignKey:VehicleTypeID"`
	Seller       User         `gorm:"foreignKey:SellerID"`
}
