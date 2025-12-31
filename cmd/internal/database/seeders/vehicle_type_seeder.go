package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedVehicleType() {

	vehicleType := []string{
		"Car/Truck",
		"Motorcycle",
		"Powersport",
		"RV/Camper",
		"Trailer",
		"Boat",
		"Commercial/Industrial",
		"Other",
	}

	for _, name := range vehicleType {
		database.DB.FirstOrCreate(
			&models.VehicleType{},
			models.VehicleType{Name: name},
		)
	}

	log.Println("Seeded vehicle_type:", vehicleType)
}
