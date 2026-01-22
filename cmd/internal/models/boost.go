package models

import (
	"time"

	"gorm.io/gorm"
)

type Boost struct {
	gorm.Model
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemType      string     `gorm:"not null" json:"item_type"` // "product" or "vehicle"
	ItemID        uint       `gorm:"not null" json:"item_id"`
	SellerID      uint       `gorm:"not null" json:"seller_id"`
	BoostType     string     `gorm:"not null" json:"boost_type"` // "premium", "featured", "top"
	DurationHours int        `gorm:"not null" json:"duration_hours"`
	Cost          float64    `gorm:"not null" json:"cost"`
	StartDate     *time.Time `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	Status        string     `gorm:"default:'pending'" json:"status"` // "pending", "active", "expired", "cancelled"

	Seller  User     `json:"seller"`
	Payment *Payment `json:"payment,omitempty"`
}

type Payment struct {
	gorm.Model
	ID            uint    `gorm:"primaryKey;autoIncrement" json:"id"`
	BoostID       uint    `json:"boost_id"`
	Amount        float64 `gorm:"not null" json:"amount"`
	Currency      string  `gorm:"default:'PHP'" json:"currency"`
	PaymentMethod string  `json:"payment_method"`
	TransactionID string  `json:"transaction_id"`
	Status        string  `gorm:"default:'pending'" json:"status"` // "pending", "completed", "failed", "refunded"

	Boost Boost `json:"boost,omitempty"`
}
