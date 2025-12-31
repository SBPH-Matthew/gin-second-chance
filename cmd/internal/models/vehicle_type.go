package models

import "gorm.io/gorm"

type VehicleType struct {
	gorm.Model
	ID   uint `gorm:"primaryKey;autoIncrement"`
	Name string
}
