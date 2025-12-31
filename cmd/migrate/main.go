package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run cmd/migrate/main.go [migrate|fresh]")
	}

	command := os.Args[1]

	database.Connect()

	switch command {
	case "migrate":
		runMigrate()
	case "fresh":
		runFresh()
	default:
		log.Fatalf("unknown command: %s", command)
	}

	log.Println("✅ Migration done")
}

func runMigrate() {
	log.Println("➡ Running migration")

	err := database.DB.AutoMigrate(
		&models.User{},
		&models.CategoryGroup{},
		&models.CategoryStatus{},
		&models.Category{},
		&models.ProductStatus{},
		&models.ProductCondition{},
		&models.Product{},
		&models.Comment{},
		&models.Region{},
		&models.Province{},
		&models.City{},
		&models.Barangay{},
		&models.VehicleType{},
		&models.Vehicle{},
	)

	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
}

func runFresh() {
	log.Println("⚠ Running FRESH migration")

	err := database.DB.Migrator().DropTable(
		&models.Comment{},
		&models.Product{},
		&models.ProductStatus{},
		&models.ProductCondition{},
		&models.Category{},
		&models.CategoryStatus{},
		&models.CategoryGroup{},
		&models.User{},
		&models.Region{},
		&models.Province{},
		&models.City{},
		&models.Barangay{},
		&models.Comment{},
		&models.VehicleType{},
		&models.Vehicle{},
	)
	if err != nil {
		log.Fatalf("drop failed: %v", err)
	}

	runMigrate()
}
