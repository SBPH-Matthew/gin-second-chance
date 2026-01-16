package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	ID              uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name            string `gorm:"unique;not null" json:"name"`
	StatusID        uint   `json:"status_id"`
	CategoryGroupID uint   `json:"category_group_id"`

	Status        CategoryStatus `json:"status"`
	CategoryGroup CategoryGroup  `json:"category_group"`
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
