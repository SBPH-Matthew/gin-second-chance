package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	ID   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"unique;not null"`

	StatusID uint

	Status CategoryStatus
}

func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.StatusID == 0 {
		var status CategoryStatus
		if err := tx.Where("name = ?", "DRAFT").First(&status).Error; err != nil {
			return err
		}
		c.StatusID = status.ID
	}
	return nil
}
