package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedCategoryStatus() {
	statuses := []models.CategoryStatus{
		{Name: "ACTIVE"},
		{Name: "INACTIVE"},
		{Name: "DRAFT"},
		{Name: "ARCHIVED"},
	}

	for _, status := range statuses {
		var existing models.CategoryStatus

		err := database.DB.
			Where("name = ?", status.Name).
			First(&existing).
			Error

		if err != nil {
			database.DB.Create(&status)
			log.Println("Seeded category_status:", status.Name)
		}
	}
}
