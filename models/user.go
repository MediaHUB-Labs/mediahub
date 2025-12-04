package models

import "gorm.io/gorm"

type User struct {
	gorm.Model        // ← Embeds 4 fields automatically
	FirstName  string `gorm:"null"`
	LastName   string `gorm:"null"`
	Email      string `gorm:"uniqueIndex;not null"`
	Password   string `gorm:"not null"`
	IsActive   bool   `gorm:"default:true"`
}
