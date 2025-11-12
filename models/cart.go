package models

import "gorm.io/gorm"

type Cart struct {
	gorm.Model
	UserID  uint       `json:"userId"`
	User    User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	OrderID *uint      `json:"orderId"`
	Order   *Order     `gorm:"foreignKey:OrderID;constraint:OnDelete:SET NULL;"`
	Items   []CartItem `json:"items" gorm:"constraint:OnDelete:CASCADE;"`
}
