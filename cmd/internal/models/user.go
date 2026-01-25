package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	FirstName      string `gorm:"not null" json:"first_name"`
	LastName       string `gorm:"not null" json:"last_name"`
	Email          string `gorm:"unique; not null" json:"email"`
	Password       string `gorm:"not null" json:"-"`
	ProfilePicture string `json:"profile_picture"`

	// Address fields
	Country        string `json:"country"`
	StateProvince  string `json:"state_province"`
	StreetAddress1 string `json:"street_address_1"`
	StreetAddress2 string `json:"street_address_2"`
	ZipPostalCode  string `json:"zip_postal_code"`

	RoleID uint `json:"role_id"`
	Role   Role `json:"role"`

	// New fields for verification and contact
	Phone            string  `json:"phone"`
	Bio              string  `gorm:"type:text" json:"bio"`
	IdentityVerified bool    `gorm:"default:false" json:"identity_verified"`
	Rating           float64 `gorm:"default:0" json:"rating"`
	TotalReviews     int     `gorm:"default:0" json:"total_reviews"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.RoleID == 0 {
		var role Role
		if err := tx.Where("name = ?", "User").First(&role).Error; err != nil {
			return err
		}
		u.RoleID = role.ID
	}
	return nil
}

func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}

	u.Password = string(bytes)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
