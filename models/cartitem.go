package models

import "gorm.io/gorm"

type CartItem struct {
	gorm.Model
	CartID    uint    `json:"cartId"`
	ProductID uint    `json:"productId"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID;constraint:OnDelete:CASCADE;"`
	Quantity  uint    `json:"quantity"`
}
