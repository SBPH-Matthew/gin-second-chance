package seeders

import (
	"log"

	"github.com/SBPH-Matthew/second-chance/cmd/internal/database"
	"github.com/SBPH-Matthew/second-chance/cmd/internal/models"
)

func SeedCategory() {
	// Map of category group names to their categories
	categoryMap := map[string][]string{
		"Home & Garden": {
			"Tools",
			"Furniture",
			"Household",
			"Garden",
			"Appliances",
		},
		"Entertainment": {
			"Video Games",
			"Books, Movies & Music",
		},
		"Clothing & Accessories": {
			"Bags & Luggage",
			"Women's clothing & shoes",
			"Men's clothing & shoes",
			"Jewelry & Accessories",
		},
		"Family": {
			"Health & beauty",
			"Pet Supplies",
			"Baby & kids",
			"Toys & Games",
		},
		"Electronics": {
			"Electronics & computers",
			"Mobile phones",
		},
		"Hobbies": {
			"Bicycles",
			"Arts & Crafts",
			"Sports & Outdoors",
			"Auto parts",
			"Musical Instruments",
			"Antiques & Collectibles",
		},
		"Vehicles": {
			"Vehicles",
		},
	}

	// Get ACTIVE status (assuming it exists from category_status_seeder)
	var activeStatus models.CategoryStatus
	if err := database.DB.Where("name = ?", "ACTIVE").First(&activeStatus).Error; err != nil {
		log.Printf("Warning: ACTIVE status not found, using default status. Error: %v\n", err)
	}

	// Iterate through each category group
	for groupName, categories := range categoryMap {
		// Find the category group
		var categoryGroup models.CategoryGroup
		if err := database.DB.Where("name = ?", groupName).First(&categoryGroup).Error; err != nil {
			log.Printf("Warning: Category group '%s' not found, skipping. Error: %v\n", groupName, err)
			continue
		}

		// Create categories for this group
		for _, categoryName := range categories {
			// Check if category already exists
			var existingCategory models.Category
			result := database.DB.Where("name = ?", categoryName).First(&existingCategory)

			if result.Error != nil {
				// Category doesn't exist, create it
				category := models.Category{
					Name:            categoryName,
					CategoryGroupID: categoryGroup.ID,
				}

				// Set status to ACTIVE if found, otherwise let BeforeCreate hook set default
				if activeStatus.ID != 0 {
					category.StatusID = activeStatus.ID
				}

				if err := database.DB.Create(&category).Error; err != nil {
					log.Printf("Error creating category '%s': %v\n", categoryName, err)
				} else {
					log.Printf("Created category: %s (Group: %s)\n", categoryName, groupName)
				}
			} else {
				// Category exists, update its group if it's different
				if existingCategory.CategoryGroupID != categoryGroup.ID {
					existingCategory.CategoryGroupID = categoryGroup.ID
					if err := database.DB.Save(&existingCategory).Error; err != nil {
						log.Printf("Error updating category '%s' group: %v\n", categoryName, err)
					} else {
						log.Printf("Updated category: %s (Group: %s)\n", categoryName, groupName)
					}
				}
			}
		}
	}

	log.Println("Category seeding completed")
}
