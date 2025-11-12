package models

import "gorm.io/gorm"

type Address struct {
	gorm.Model
	UserID       uint   `json:"userId"`
	AddressPlace string `json:"addressplace"`
	Postalcode   uint   `json:"postalcode"`
	PhoneNumber  string `json:"phonenumber"`
}
