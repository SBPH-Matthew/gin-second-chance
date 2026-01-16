package models

import "gorm.io/gorm"

type CategoryStatus struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}
