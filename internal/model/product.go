// Package model contains the database entities (GORM models).
package model

import (
	"time" // time is used for the timestamp fields

	"gorm.io/gorm" // gorm provides the DeletedAt type for soft deletes
)

// Product is the database representation of a product (table: products).
type Product struct {
	ID               uint           `gorm:"primaryKey"`                  // ID is the auto-increment primary key
	Name             string         `gorm:"type:varchar(255);not null"`  // Name of the product (required)
	Price            float64        `gorm:"type:numeric(12,2);not null"` // Price stored with 2 decimal precision
	ShortDescription string         `gorm:"type:varchar(500)"`           // ShortDescription is a brief summary
	CreatedAt        time.Time      // CreatedAt is set automatically by GORM on insert
	UpdatedAt        time.Time      // UpdatedAt is set automatically by GORM on update
	DeletedAt        gorm.DeletedAt `gorm:"index"` // DeletedAt enables soft delete
}

// TableName tells GORM to use the "products" table name explicitly.
func (Product) TableName() string {
	// Return the table name
	return "products"
}
