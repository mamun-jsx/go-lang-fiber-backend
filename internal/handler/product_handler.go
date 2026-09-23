// Package handler is the HTTP (transport) layer: it parses requests and writes responses.
// Handlers contain NO business logic and NO database code.
package handler

import (
	"net/http" // net/http provides HTTP status codes
	"strconv"  // strconv parses the :id path parameter

	"github.com/gofiber/fiber/v3" // fiber is the web framework

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/apperror" // typed app errors
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/dto"      // request/response shapes
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/service"  // business logic layer
	"github.com/mamun-jsx/go-lang-fiber-backend/pkg/response"      // JSON envelope helpers
	"github.com/mamun-jsx/go-lang-fiber-backend/pkg/validator"     // request validation
)

// ProductHandler groups all HTTP handlers for the product resource.
type ProductHandler struct {
	service   service.ProductService // service is the business logic dependency
	validator *validator.Validator   // validator validates request bodies
}

// NewProductHandler creates a handler with its dependencies injected.
func NewProductHandler(s service.ProductService, v *validator.Validator) *ProductHandler {
	// Return the handler instance
	return &ProductHandler{service: s, validator: v}
}

// Create handles POST /api/v1/products.
func (h *ProductHandler) Create(c fiber.Ctx) error {
	// Declare the request DTO to bind the JSON body into
	var req dto.CreateProductRequest
	// Parse the JSON body; malformed JSON -> 400
	if err := c.Bind().JSON(&req); err != nil {
		return apperror.BadRequest("invalid JSON body")
	}
	// Validate the struct tags; invalid fields -> 422
	if errs := h.validator.Validate(req); errs != nil {
		return apperror.Validation(errs)
	}
	// Call the service layer with the request context
	product, err := h.service.Create(c.Context(), req)
	// Let the global error handler format any error
	if err != nil {
		return err
	}
	// Respond with 201 Created and the new product
	return response.Success(c, http.StatusCreated, "product created successfully", product)
}

// List handles GET /api/v1/products?page=1&limit=10.
func (h *ProductHandler) List(c fiber.Ctx) error {
	// Declare the pagination query DTO
	var q dto.PaginationQuery
	// Bind query string parameters (?page=&limit=)
	if err := c.Bind().Query(&q); err != nil {
		return apperror.BadRequest("invalid query parameters")
	}
	// Call the service to fetch the page
	products, meta, err := h.service.List(c.Context(), q)
	// Forward errors to the global error handler
	if err != nil {
		return err
	}
	// Respond with 200 OK, the list and pagination meta
	return response.SuccessWithMeta(c, http.StatusOK, "products fetched successfully", products, meta)
}

// GetByID handles GET /api/v1/products/:id.
func (h *ProductHandler) GetByID(c fiber.Ctx) error {
	// Parse and validate the :id path parameter
	id, err := parseID(c)
	// Invalid ID -> 400
	if err != nil {
		return err
	}
	// Ask the service for the product
	product, err := h.service.GetByID(c.Context(), id)
	// Forward errors (e.g. 404) to the global error handler
	if err != nil {
		return err
	}
	// Respond with 200 OK and the product
	return response.Success(c, http.StatusOK, "product fetched successfully", product)
}

// Update handles PUT /api/v1/products/:id (partial update: send only fields to change).
func (h *ProductHandler) Update(c fiber.Ctx) error {
	// Parse and validate the :id path parameter
	id, err := parseID(c)
	// Invalid ID -> 400
	if err != nil {
		return err
	}
	// Declare the update DTO
	var req dto.UpdateProductRequest
	// Parse the JSON body; malformed JSON -> 400
	if err := c.Bind().JSON(&req); err != nil {
		return apperror.BadRequest("invalid JSON body")
	}
	// Validate the provided fields -> 422 on failure
	if errs := h.validator.Validate(req); errs != nil {
		return apperror.Validation(errs)
	}
	// Call the service to apply the update
	product, err := h.service.Update(c.Context(), id, req)
	// Forward errors to the global error handler
	if err != nil {
		return err
	}
	// Respond with 200 OK and the updated product
	return response.Success(c, http.StatusOK, "product updated successfully", product)
}

// Delete handles DELETE /api/v1/products/:id.
func (h *ProductHandler) Delete(c fiber.Ctx) error {
	// Parse and validate the :id path parameter
	id, err := parseID(c)
	// Invalid ID -> 400
	if err != nil {
		return err
	}
	// Ask the service to delete the product
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	// Respond with 200 OK and a confirmation message
	return response.Success(c, http.StatusOK, "product deleted successfully", nil)
}

// parseID reads the :id route param and ensures it is a positive integer.
func parseID(c fiber.Ctx) (uint, error) {
	// Convert the string param into an unsigned integer
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	// Reject non-numeric or zero IDs
	if err != nil || id == 0 {
		return 0, apperror.BadRequest("invalid product id")
	}
	// Return the parsed ID as uint
	return uint(id), nil
}
