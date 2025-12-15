package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"not null"`
	Category string `gorm:"not null"`
}
