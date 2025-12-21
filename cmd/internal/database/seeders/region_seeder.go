package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/utils"
)

type RegionJSON struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	RegionName      string `json:"regionName"`
	IslandGroupCode string `json:"islandGroupCode"`
	Psgc10DigitCode string `json:"psgc10DigitCode"`
}

func SeedRegions() error {
	regionsJSON, err := utils.LoadJSON[RegionJSON]("data/psgc/regions.json")
	if err != nil {
		return err
	}

	var regions []models.Region
	for _, r := range regionsJSON {
		regions = append(regions, models.Region{
			Code:            r.Code,
			Name:            r.Name,
			RegionName:      r.RegionName,
			IslandGroupCode: r.IslandGroupCode,
			Psgc10DigitCode: r.Psgc10DigitCode,
		})
	}

	if err := database.DB.CreateInBatches(&regions, 1000).Error; err != nil {
		return err
	}

	log.Println("Regions seeded successfully")
	return nil
}
