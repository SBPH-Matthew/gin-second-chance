package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
)

type CityJSON struct {
	Code           string `json:"code"`
	Name           string `json:"name"`
	OldName        string `json:"oldName"`
	IsCapital      bool   `json:"isCapital"`
	IsCity         bool   `json:"isCity"`
	IsMunicipality bool   `json:"isMunicipality"`

	ProvinceCode    utils.NullableString `json:"provinceCode"`
	DistrictCode    utils.NullableString `json:"districtCode"`
	RegionCode      utils.NullableString `json:"regionCode"`
	IslandGroupCode utils.NullableString `json:"islandGroupCode"`

	Psgc10DigitCode string `json:"psgc10DigitCode"`
}

func SeedCity() error {
	cityJSON, err := utils.LoadJSON[CityJSON]("data/psgc/cities.json")
	if err != nil {
		log.Fatalf("failed to load city JSON: %v", err)
		return err
	}

	var cities []models.City
	for _, city := range cityJSON {
		cities = append(cities, models.City{
			Code:            city.Code,
			Name:            city.Name,
			OldName:         city.OldName,
			IsCapital:       city.IsCapital,
			IsCity:          city.IsCity,
			IsMunicipality:  city.IsMunicipality,
			ProvinceCode:    city.ProvinceCode.Value,
			DistrictCode:    city.DistrictCode.Value,
			RegionCode:      city.RegionCode.Value,
			IslandGroupCode: city.IslandGroupCode.Value,
			Psgc10DigitCode: city.Psgc10DigitCode,
		})
	}

	if err := database.DB.CreateInBatches(&cities, 1000).Error; err != nil {
		log.Fatalf("failed to seed cities: %v", err)
		return err
	}

	log.Println("Cities seeded successfully")
	return nil
}
