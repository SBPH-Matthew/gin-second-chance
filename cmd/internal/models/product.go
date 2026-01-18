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
	ID                uint        `gorm:"primaryKey;autoIncrement" json:"id"`
	Name              string      `gorm:"not null" json:"name"`
	Description       string      `gorm:"not null" json:"description"`
	Price             float64     `gorm:"not null" json:"price"`
	Location          string      `gorm:"not null" json:"location"`
	Images            StringArray `gorm:"type:json" json:"images"`
	CategoryID        uint        `json:"category_id"`
	StatusID          uint        `json:"status_id"`
	SellerID          uint        `json:"seller_id"`
	ProductConditionID uint       `json:"product_condition_id"`

	Category         Category         `json:"category"`
	Status           ProductStatus    `json:"status"`
	Seller           User             `json:"seller"`
	ProductCondition ProductCondition `json:"product_condition"`
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
