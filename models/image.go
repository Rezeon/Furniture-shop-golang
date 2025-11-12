package models

import "gorm.io/gorm"

type Image struct {
	gorm.Model

	Image string `json:"image"`
	Name  string `json:"name"`
}
