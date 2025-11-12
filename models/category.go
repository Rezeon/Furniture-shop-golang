package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name     string    `json:"name"`
	Products []Product `json:"products" gorm:"foreignKey:CategoryID;constraint:OnDelete:CASCADE;"`
}
