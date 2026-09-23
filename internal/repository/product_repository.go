// Package repository is the data-access layer: the ONLY layer that talks to the database.
package repository

import (
	"context" // context carries deadlines / cancellation into DB queries
	"errors"  // errors is used for errors.Is

	"gorm.io/gorm" // gorm is the ORM used to run queries

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/model" // model is the DB entity
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// ProductRepository defines the data operations available for products.
// Depending on an interface (not a struct) makes the service easy to unit-test.
type ProductRepository interface {
	Create(ctx context.Context, p *model.Product) error                             // insert a product
	FindAll(ctx context.Context, offset, limit int) ([]model.Product, int64, error) // list with pagination
	FindByID(ctx context.Context, id uint) (*model.Product, error)                  // get one by ID
	Update(ctx context.Context, p *model.Product) error                             // save changes
	Delete(ctx context.Context, id uint) error                                      // soft delete by ID
}

// productRepository is the GORM implementation of ProductRepository.
type productRepository struct {
	db *gorm.DB // db is the GORM connection
}

// NewProductRepository creates a new repository bound to the given DB connection.
func NewProductRepository(db *gorm.DB) ProductRepository {
	// Return the implementation as the interface type
	return &productRepository{db: db}
}

// Create inserts a new product row; GORM fills ID, CreatedAt and UpdatedAt.
func (r *productRepository) Create(ctx context.Context, p *model.Product) error {
	// INSERT INTO products (...) VALUES (...)
	return r.db.WithContext(ctx).Create(p).Error
}

// FindAll returns one page of products plus the total number of products.
func (r *productRepository) FindAll(ctx context.Context, offset, limit int) ([]model.Product, int64, error) {
	var (
		products []model.Product // products will hold the current page
		total    int64           // total will hold the count of all rows
	)
	// Build a base query bound to the request context
	q := r.db.WithContext(ctx).Model(&model.Product{})
	// SELECT count(*) FROM products WHERE deleted_at IS NULL
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	// SELECT * FROM products ORDER BY id DESC LIMIT ? OFFSET ?
	if err := q.Order("id DESC").Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, 0, err
	}
	// Return the page and the total count
	return products, total, nil
}

// FindByID returns a single product or ErrNotFound.
func (r *productRepository) FindByID(ctx context.Context, id uint) (*model.Product, error) {
	// Variable that GORM will fill with the row
	var p model.Product
	// SELECT * FROM products WHERE id = ? LIMIT 1
	err := r.db.WithContext(ctx).First(&p, id).Error
	// Translate GORM's not-found error into our own repository error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	// Any other DB error is returned as-is
	if err != nil {
		return nil, err
	}
	// Return the found product
	return &p, nil
}

// Update saves all fields of an existing product.
func (r *productRepository) Update(ctx context.Context, p *model.Product) error {
	// UPDATE products SET ... WHERE id = ?
	return r.db.WithContext(ctx).Save(p).Error
}

// Delete soft-deletes a product (sets deleted_at) or returns ErrNotFound.
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	// UPDATE products SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL
	res := r.db.WithContext(ctx).Delete(&model.Product{}, id)
	// Return DB errors
	if res.Error != nil {
		return res.Error
	}
	// No rows affected means the product did not exist
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	// Deleted successfully
	return nil
}
