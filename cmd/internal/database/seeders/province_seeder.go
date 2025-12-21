package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
)

type ProvinceJSON struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	RegionCode      string `json:"region_code"`
	IslandGroupCode string `json:"island_group_code"`
	Psgc10DigitCode string `json:"psgc10DigitCode"`
}

func SeedProvinces() error {

	provinceJSON, err := utils.LoadJSON[ProvinceJSON]("data/psgc/provinces.json")
	if err != nil {
		return err
	}

	var province []models.Province
	for _, p := range provinceJSON {
		province = append(province, models.Province{
			Code:            p.Code,
			Name:            p.Name,
			RegionCode:      p.RegionCode,
			IslandGroupCode: p.IslandGroupCode,
			Psgc10DigitCode: p.Psgc10DigitCode,
		})
	}

	if err := database.DB.CreateInBatches(&province, 1000).Error; err != nil {
		return err
	}

	log.Println("Provinces seeded successfully")
	return nil
}
