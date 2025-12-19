package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedProductCondition() {

	product_condition := []string{"New", "Used - Like New", "Used - Good", "Used - Fair", "Used - Poor"}

	for _, name := range product_condition {
		database.DB.FirstOrCreate(
			&models.ProductCondition{},
			models.ProductCondition{Name: name},
		)
	}

	log.Println("Seeded product_condition:", product_condition)
}
