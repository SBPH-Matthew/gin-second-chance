package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"unique;not null"`

	StatusID        uint
	CategoryGroupID uint

	Status        CategoryStatus
	CategoryGroup CategoryGroup
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.StatusID == 0 {
		var status CategoryStatus
		if err := tx.Where("name = ?", "DRAFT").First(&status).Error; err != nil {
			return err
		}
		c.StatusID = status.ID
	}

	if c.CategoryGroupID == 0 {
		var categoryGroup CategoryGroup
		if err := tx.Where("name = ?", "Others").First(&categoryGroup).Error; err != nil {
			return err
		}
		c.CategoryGroupID = categoryGroup.ID
	}

	return nil
}
