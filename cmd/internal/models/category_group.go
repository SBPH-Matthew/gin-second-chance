package models

import "gorm.io/gorm"

type CategoryGroup struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"not null;unique;size:255"`
}
