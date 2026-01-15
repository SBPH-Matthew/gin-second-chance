package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

// StringArray is a custom type for storing JSON array of strings
type StringArray []string

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal StringArray value")
	}

	return json.Unmarshal(bytes, a)
}

type Product struct {
	gorm.Model
	ID          uint        `gorm:"primaryKey;autoIncrement"`
	Name        string      `gorm:"not null"`
	Description string      `gorm:"not null"`
	Price       float64     `gorm:"not null"`
	Images      StringArray `gorm:"type:json"`

	CategoryID         uint
	StatusID           uint
	SellerID           uint
	ProductConditionID uint

	Category         Category
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
