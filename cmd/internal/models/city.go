package models

import "gorm.io/gorm"

type City struct {
	gorm.Model
	ID              uint   `gorm:"primaryKey"`
	Code            string `gorm:"uniqueIndex;size:9"`
	Name            string
	OldName         string
	IsCapital       bool
	IsCity          bool
	IsMunicipality  bool
	ProvinceCode    string `gorm:"index"`
	DistrictCode    string `gorm:"index"`
	RegionCode      string `gorm:"index"`
	IslandGroupCode string `gorm:"size:10;index"`
	Psgc10DigitCode string `gorm:"size:10;index"`
}
