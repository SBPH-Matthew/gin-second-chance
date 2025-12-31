package models

import "gorm.io/gorm"

type Vehicle struct {
	gorm.Model

	VehicleTypeID *uint
	Location      string

	Year         uint
	VehicleMake  string
	VehicleModel string
	Price        uint
	Description  string
	VehicleType  VehicleType `gorm:"foreignKey:VehicleTypeID"`
}
