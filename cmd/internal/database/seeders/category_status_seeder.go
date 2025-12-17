package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedCategoryStatus() {

	statuses := []string{"ACTIVE", "INACTIVE", "DRAFT", "ARCHIVED"}

	for _, name := range statuses {
		database.DB.FirstOrCreate(
			&models.CategoryStatus{},
			models.CategoryStatus{Name: name},
		)
	}

	log.Println("Seeded category_status:", statuses)
}
