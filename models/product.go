package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	Name        string     `json:"name"`
	Price       uint       `json:"price"`
	Image       string     `json:"image"`
	PublicID    string     `json:"public_id"`
	Description string     `json:"description"`
	CategoryID  uint       `json:"categoryId"`
	Category    Category   `json:"category" gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;"`
	CartItems   []CartItem `json:"cartItems" gorm:"constraint:OnDelete:CASCADE;"`
}
