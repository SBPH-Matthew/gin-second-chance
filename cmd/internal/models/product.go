package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID          uint    `gorm:"primaryKey,autoIncrement"`
	Name        string  `gorm:"not null"`
	Description string  `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	CategoryID  uint
	StatusID    uint
	SellerID    uint

	CategoryRefer Category      `gorm:"foreignKey:CategoryID"`
	StatusRefer   ProductStatus `gorm:"foreignKey:StatusID"`
	SellerRefer   User          `gorm:"foreignKey:SellerID"`
}
