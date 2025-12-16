package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey,autoIncrement"`
	Name        string `gorm:"not null"`
	Status      uint
	StatusRefer CategoryStatus `gorm:"foreignKey:Status"`
}
