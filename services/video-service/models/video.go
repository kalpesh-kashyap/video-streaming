package models

import "gorm.io/gorm"

type Video struct {
	gorm.Model
	Title       string `gorm:"size:100;not null"`
	Description string `gorm:"size: 300"`
	Size        int    `gorm:"not null"`
}
