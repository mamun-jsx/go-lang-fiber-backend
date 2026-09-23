// Package service contains the business logic. It sits between handlers and repositories.
package service

import (
	"context" // context is passed through to the repository
	"errors"  // errors is used for errors.Is
	"strings" // strings is used to trim input

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/apperror"   // typed app errors
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/dto"        // request/response shapes
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/model"      // DB entity
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/repository" // data access layer
)

// ProductService defines the product use-cases exposed to the handler layer.
type ProductService interface {
	Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)             // create
	List(ctx context.Context, q dto.PaginationQuery) ([]dto.ProductResponse, dto.PaginationMeta, error) // list
	GetByID(ctx context.Context, id uint) (*dto.ProductResponse, error)                                 // read one
	Update(ctx context.Context, id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error)    // update
	Delete(ctx context.Context, id uint) error                                                          // delete
}

// productService is the default implementation of ProductService.
type productService struct {
	repo repository.ProductRepository // repo is the data layer dependency (interface)
}

// NewProductService creates a service with its repository dependency injected.
func NewProductService(repo repository.ProductRepository) ProductService {
	// Return the implementation as the interface type
	return &productService{repo: repo}
}

// Create builds a new product from the request and stores it.
func (s *productService) Create(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Map the DTO to a DB model, trimming surrounding whitespace
	p := &model.Product{
		Name:             strings.TrimSpace(req.Name),
		Price:            req.Price,
		ShortDescription: strings.TrimSpace(req.ShortDescription),
	}
	// Ask the repository to persist the product
	if err := s.repo.Create(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	// Convert the saved model into a response DTO
	res := dto.ToProductResponse(p)
	// Return the created product
	return &res, nil
}

// List returns a paginated list of products.
func (s *productService) List(ctx context.Context, q dto.PaginationQuery) ([]dto.ProductResponse, dto.PaginationMeta, error) {
	// Apply default / max values to page and limit
	q.Normalize()
	// Fetch one page and the total count from the repository
	products, total, err := s.repo.FindAll(ctx, q.Offset(), q.Limit)
	// Wrap DB errors as internal errors
	if err != nil {
		return nil, dto.PaginationMeta{}, apperror.Internal(err)
	}
	// Convert models to DTOs and build the pagination meta
	return dto.ToProductResponseList(products), dto.NewPaginationMeta(q, total), nil
}

// GetByID returns one product or a 404 error.
func (s *productService) GetByID(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	// Load the product (shared helper handles not-found mapping)
	p, err := s.findOrFail(ctx, id)
	// Propagate the error
	if err != nil {
		return nil, err
	}
	// Convert to DTO
	res := dto.ToProductResponse(p)
	// Return the product
	return &res, nil
}

// Update changes only the fields provided in the request (partial update).
func (s *productService) Update(ctx context.Context, id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	// Business rule: at least one field must be provided
	if req.Name == nil && req.Price == nil && req.ShortDescription == nil {
		return nil, apperror.BadRequest("at least one field must be provided")
	}
	// Load the existing product first
	p, err := s.findOrFail(ctx, id)
	// Propagate not-found / internal errors
	if err != nil {
		return nil, err
	}
	// Apply the new name if it was provided
	if req.Name != nil {
		p.Name = strings.TrimSpace(*req.Name)
	}
	// Apply the new price if it was provided
	if req.Price != nil {
		p.Price = *req.Price
	}
	// Apply the new description if it was provided
	if req.ShortDescription != nil {
		p.ShortDescription = strings.TrimSpace(*req.ShortDescription)
	}
	// Persist the changes
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperror.Internal(err)
	}
	// Convert to DTO
	res := dto.ToProductResponse(p)
	// Return the updated product
	return &res, nil
}

// Delete removes a product or returns a 404 error.
func (s *productService) Delete(ctx context.Context, id uint) error {
	// Ask the repository to delete the product
	err := s.repo.Delete(ctx, id)
	// Map repository not-found to a 404 application error
	if errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("product not found")
	}
	// Map any other DB error to a 500 application error
	if err != nil {
		return apperror.Internal(err)
	}
	// Deleted successfully
	return nil
}

// findOrFail loads a product and translates repository errors into app errors.
func (s *productService) findOrFail(ctx context.Context, id uint) (*model.Product, error) {
	// Query the repository
	p, err := s.repo.FindByID(ctx, id)
	// Not found -> 404
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("product not found")
	}
	// Other errors -> 500
	if err != nil {
		return nil, apperror.Internal(err)
	}
	// Found
	return p, nil
}
