package models

import "gorm.io/gorm"

type ProductStatus struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null,unique"`
}
