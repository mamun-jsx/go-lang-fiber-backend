// Package dto defines Data Transfer Objects: the request/response shapes of the API.
// Keeping them separate from models means the DB schema never leaks to clients.
package dto

import (
	"time" // time is used for timestamp fields in responses

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/model" // model is converted to DTOs
)

// CreateProductRequest is the JSON body for POST /api/v1/products.
type CreateProductRequest struct {
	Name             string  `json:"name" validate:"required,min=2,max=255"`         // Name is required, 2..255 chars
	Price            float64 `json:"price" validate:"required,gt=0"`                 // Price must be greater than 0
	ShortDescription string  `json:"short_description" validate:"omitempty,max=500"` // optional, max 500 chars
}

// UpdateProductRequest is the JSON body for PUT /api/v1/products/:id.
// Pointer fields make every field optional: nil means "do not change".
type UpdateProductRequest struct {
	Name             *string  `json:"name" validate:"omitempty,min=2,max=255"`        // optional new name
	Price            *float64 `json:"price" validate:"omitempty,gt=0"`                // optional new price
	ShortDescription *string  `json:"short_description" validate:"omitempty,max=500"` // optional new description
}

// ProductResponse is the JSON shape returned to clients for a product.
type ProductResponse struct {
	ID               uint      `json:"id"`                // ID of the product
	Name             string    `json:"name"`              // Name of the product
	Price            float64   `json:"price"`             // Price of the product
	ShortDescription string    `json:"short_description"` // ShortDescription of the product
	CreatedAt        time.Time `json:"created_at"`        // CreatedAt timestamp
	UpdatedAt        time.Time `json:"updated_at"`        // UpdatedAt timestamp
}

// PaginationQuery holds the ?page=&limit= query parameters for list endpoints.
type PaginationQuery struct {
	Page  int `query:"page"`  // Page number, starting from 1
	Limit int `query:"limit"` // Limit is the number of items per page
}

// Normalize applies defaults and bounds to the pagination values.
func (p *PaginationQuery) Normalize() {
	// Page must be at least 1
	if p.Page < 1 {
		p.Page = 1
	}
	// Default limit is 10 when missing or invalid
	if p.Limit < 1 {
		p.Limit = 10
	}
	// Cap the limit to protect the database from huge queries
	if p.Limit > 100 {
		p.Limit = 100
	}
}

// Offset returns the number of rows to skip for the current page.
func (p PaginationQuery) Offset() int {
	// (page - 1) * limit rows are skipped
	return (p.Page - 1) * p.Limit
}

// PaginationMeta describes pagination info returned alongside list data.
type PaginationMeta struct {
	Page       int   `json:"page"`        // Page is the current page
	Limit      int   `json:"limit"`       // Limit is the page size
	Total      int64 `json:"total"`       // Total is the total number of records
	TotalPages int   `json:"total_pages"` // TotalPages is the total number of pages
}

// NewPaginationMeta builds pagination metadata from the query and total count.
func NewPaginationMeta(q PaginationQuery, total int64) PaginationMeta {
	// Compute total pages using ceiling division
	totalPages := int((total + int64(q.Limit) - 1) / int64(q.Limit))
	// Return the populated meta struct
	return PaginationMeta{Page: q.Page, Limit: q.Limit, Total: total, TotalPages: totalPages}
}

// ToProductResponse converts a DB model into an API response DTO.
func ToProductResponse(p *model.Product) ProductResponse {
	// Copy only the fields that should be exposed to clients
	return ProductResponse{
		ID:               p.ID,
		Name:             p.Name,
		Price:            p.Price,
		ShortDescription: p.ShortDescription,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

// ToProductResponseList converts a slice of models into a slice of DTOs.
func ToProductResponseList(products []model.Product) []ProductResponse {
	// Pre-allocate the result slice with the right capacity
	out := make([]ProductResponse, 0, len(products))
	// Convert each product one by one
	for i := range products {
		out = append(out, ToProductResponse(&products[i]))
	}
	// Return the converted list
	return out
}
