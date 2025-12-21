package models

import "gorm.io/gorm"

type Province struct {
	gorm.Model
	ID              uint   `gorm:"primaryKey"`
	Code            string `gorm:"uniqueIndex;size:9"`
	Name            string
	RegionCode      string
	IslandGroupCode string `gorm:"size:10;index"`
	Psgc10DigitCode string `gorm:"size:10;index"`
}
