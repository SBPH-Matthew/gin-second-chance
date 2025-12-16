package models

import "gorm.io/gorm"

type CategoryStatus struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null"`
}
