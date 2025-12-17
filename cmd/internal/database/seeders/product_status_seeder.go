package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedProductStatus() {
	statuses := []string{"ACTIVE", "INACTIVE", "DRAFT", "ARCHIVED"}

	for _, name := range statuses {
		database.DB.FirstOrCreate(
			&models.ProductStatus{},
			models.ProductStatus{Name: name},
		)
	}

	log.Println("Seeded product_status:", statuses)
}
