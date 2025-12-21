package models

import "gorm.io/gorm"

type Barangay struct {
	gorm.Model
	ID               uint   `gorm:"primaryKey"`
	Code             string `gorm:"uniqueIndex;size:9"`
	Name             string
	OldName          string
	SubMunicipality  string
	CityCode         string
	MunicipalityCode string
	DistrictCode     string
	ProvinceCode     string
	RegionCode       string
	IslandGroupCode  string
	Psgc10DigitCode  string
}
