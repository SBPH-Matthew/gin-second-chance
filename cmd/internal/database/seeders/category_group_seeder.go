package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedCategoryGroup() {

	category_group := []string{"Home & Garden", "Entertainment", "Clothing & Accessories", "Family", "Electronics", "Hobbies", "Classifieds", "Vehicles", "Others"}

	for _, name := range category_group {
		database.DB.FirstOrCreate(
			&models.CategoryGroup{},
			models.CategoryGroup{Name: name},
		)
	}

	log.Println("Seeded category_group:", category_group)
}
