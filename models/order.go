package models

import "gorm.io/gorm"

type Order struct {
	gorm.Model
	UserID          uint   `json:"userId"`
	Cart            []Cart `json:"cart" gorm:"constraint:OnDelete:CASCADE;"`
	TotalPrice      uint   `json:"totalPrice"`
	Payment         string `json:"payment"`
	Quantity        uint   `json:"quantity"`
	Status          string `json:"status" gorm:"default:'Pending'"`
	DuitkuReference string `json:"duitkuReference"`
}

//type Order struct {
//    gorm.Model
//    UserID     uint    `json:"userId"`
//    Cart       []Cart  `json:"cart" gorm:"constraint:OnDelete:CASCADE;"`
//    TotalPrice uint    `json:"totalPrice"`
//    Payment    string  `json:"payment"`
//    Quantity   uint    `json:"quantity"`
//    Status     string  `json:"status" gorm:"default:'Pending'"`
//    DuitkuReference string `json:"duitkuReference"`
//}
