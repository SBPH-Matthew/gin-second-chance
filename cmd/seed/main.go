package main

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/database/seeders"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	database.Connect()

	log.Println("Running seeders...")

	seeders.SeedCategoryStatus()
	seeders.SeedCategoryGroup()
	seeders.SeedProductStatus()
	seeders.SeedProductCondition()
	seeders.SeedRegions()
	seeders.SeedProvinces()
	seeders.SeedCity()
	seeders.SeedBarangays()
	seeders.SeedRoles()

	log.Println("Seeding completed!")
}
