package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	User      User `gorm:"foreignKey:UserID"`
	ProductID uint
	Product   Product `gorm:"foreignKey:ProductID"`
	Content   string
}
