package models

import "gorm.io/gorm"

type ContactViewLog struct {
	gorm.Model
	ViewerID  uint    `gorm:"not null" json:"viewer_id"`
	SellerID  uint    `gorm:"not null" json:"seller_id"`
	ProductID uint    `gorm:"not null" json:"product_id"`
	Viewer    User    `json:"viewer"`
	Seller    User    `json:"seller"`
	Product   Product `json:"product"`
}
