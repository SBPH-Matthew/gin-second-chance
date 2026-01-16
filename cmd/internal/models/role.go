package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"not null;unique" json:"name"`
}
