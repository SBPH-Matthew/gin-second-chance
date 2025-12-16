package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedProductStatus() {
	statuses := []models.ProductStatus{
		{Name: "ACTIVE"},
		{Name: "INACTIVE"},
		{Name: "DRAFT"},
		{Name: "ARCHIVED"},
	}

	for _, status := range statuses {
		var existing models.ProductStatus

		err := database.DB.
			Where("name = ?", status.Name).
			First(&existing).
			Error

		if err != nil {
			database.DB.Create(&status)
			log.Println("Seeded product_status:", status.Name)
		}
	}
}
