package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Password  string  `json:"-"`
	Role      string  `json:"role"`
	AddressID *uint   `json:"addressId"`
	Address   Address `gorm:"foreignKey:AddressID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Orders    []Order `json:"orders" gorm:"constraint:OnDelete:CASCADE;"`
	Carts     []Cart  `json:"carts" gorm:"constraint:OnDelete:CASCADE;"`
}
