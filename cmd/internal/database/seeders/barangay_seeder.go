package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
)

type BarangayJSON struct {
	Code             string               `json:"code"`
	Name             string               `json:"name"`
	OldName          string               `json:"oldName"`
	SubMunicipality  utils.NullableString `json:"subMunicipality"`
	CityCode         utils.NullableString `json:"cityCode"`
	MunicipalityCode utils.NullableString `json:"municipalityCode"`
	DistrictCode     utils.NullableString `json:"districtCode"`
	ProvinceCode     utils.NullableString `json:"provinceCode"`
	RegionCode       utils.NullableString `json:"regionCode"`
	IslandGroupCode  utils.NullableString `json:"islandGroupCode"`
	Psgc10DigitCode  string               `json:"psgc10DigitCode"`
}

func SeedBarangays() error {
	barangayJSON, err := utils.LoadJSON[BarangayJSON]("data/psgc/barangays.json")
	if err != nil {
		return err
	}

	var barangays []models.Barangay
	for _, barangay := range barangayJSON {
		barangays = append(barangays, models.Barangay{
			Code:             barangay.Code,
			Name:             barangay.Name,
			OldName:          barangay.OldName,
			SubMunicipality:  barangay.SubMunicipality.Value,
			CityCode:         barangay.CityCode.Value,
			MunicipalityCode: barangay.MunicipalityCode.Value,
			DistrictCode:     barangay.DistrictCode.Value,
			ProvinceCode:     barangay.ProvinceCode.Value,
			RegionCode:       barangay.RegionCode.Value,
			IslandGroupCode:  barangay.IslandGroupCode.Value,
			Psgc10DigitCode:  barangay.Psgc10DigitCode,
		})
	}

	if err := database.DB.CreateInBatches(&barangays, 1000).Error; err != nil {
		log.Fatal(err)
		return err
	}

	log.Println("Barangays seeded successfully.")
	return nil
}
