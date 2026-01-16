package models

import "gorm.io/gorm"

type VehicleType struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `json:"name"`
}
