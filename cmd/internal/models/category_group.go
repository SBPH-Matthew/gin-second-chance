package models

import "gorm.io/gorm"

type CategoryGroup struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null;unique;size:255" json:"name"`
}
