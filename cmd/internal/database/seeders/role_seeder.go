package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedRoles() {

	roles := []string{"admin", "user", "staff"}

	for _, role := range roles {
		database.DB.FirstOrCreate(
			&models.Role{},
			models.Role{Name: role},
		)
	}

	log.Println("Roles seeded successfully")
}
