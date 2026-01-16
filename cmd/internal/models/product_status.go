package models

import "gorm.io/gorm"

type ProductStatus struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
}
