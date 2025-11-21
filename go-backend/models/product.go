package models

import (
	"gorm.io/gorm"
)

// Product represents a single item in the inventory.
// This is the source of truth for the 'products' table schema.
type Product struct {
	// gorm.Model includes common fields: ID, CreatedAt, UpdatedAt, DeletedAt
	gorm.Model
	Name  string  `json:"name" gorm:"not null"`
	Price float64 `json:"price" gorm:"type:numeric"`
}
