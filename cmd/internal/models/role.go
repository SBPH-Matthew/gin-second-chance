package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;unique"`
}
