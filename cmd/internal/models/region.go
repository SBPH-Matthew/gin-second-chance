package models

import "gorm.io/gorm"

type Region struct {
	gorm.Model
	ID              uint   `gorm:"primaryKey"`
	Code            string `gorm:"size:9;uniqueIndex"`
	Name            string
	RegionName      string
	IslandGroupCode string `gorm:"size:10;index"`
	Psgc10DigitCode string `gorm:"size:10;index"`
}
