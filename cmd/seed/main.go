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
	seeders.SeedProductStatus()

	log.Println("Seeding completed!")
}
