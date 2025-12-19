package models

import (
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID          uint    `gorm:"primaryKey;autoIncrement"`
	Name        string  `gorm:"not null"`
	Description string  `gorm:"not null"`
	Price       float64 `gorm:"not null"`

	CategoryID         *uint
	StatusID           uint
	SellerID           uint
	ProductConditionID uint

	Category         *Category
	Status           ProductStatus
	Seller           User
	ProductCondition ProductCondition
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.StatusID == 0 {
		var status ProductStatus
		if err := tx.Where("name = ?", "DRAFT").First(&status).Error; err != nil {
			return err
		}
		p.StatusID = status.ID
	}

	if p.ProductConditionID == 0 {
		var condition ProductCondition
		if err := tx.Where("name = ?", "New").First(&condition).Error; err != nil {
			return err
		}
		p.ProductConditionID = condition.ID
	}

	return nil
}
