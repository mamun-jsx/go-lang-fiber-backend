package service

import (
	"context"  // context is passed to service methods
	"net/http" // net/http provides expected status codes
	"testing"  // testing is Go's built-in test framework

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/apperror"   // to inspect returned errors
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/dto"        // request DTOs
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/model"      // DB entity
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/repository" // ErrNotFound
)

// fakeRepo is an in-memory ProductRepository used to test the service without a database.
type fakeRepo struct {
	data   map[uint]*model.Product // data stores products by ID
	nextID uint                    // nextID simulates auto-increment
}

// newFakeRepo creates an empty in-memory repository.
func newFakeRepo() *fakeRepo { return &fakeRepo{data: map[uint]*model.Product{}, nextID: 1} }

// Create stores the product and assigns an ID.
func (f *fakeRepo) Create(_ context.Context, p *model.Product) error {
	p.ID = f.nextID  // assign the next ID
	f.nextID++       // increment for the next insert
	f.data[p.ID] = p // store the product
	return nil       // no error
}

// FindAll returns all stored products (pagination ignored for simplicity).
func (f *fakeRepo) FindAll(_ context.Context, _, _ int) ([]model.Product, int64, error) {
	out := []model.Product{}   // result slice
	for _, p := range f.data { // copy every product
		out = append(out, *p)
	}
	return out, int64(len(out)), nil // return list and count
}

// FindByID returns the product or repository.ErrNotFound.
func (f *fakeRepo) FindByID(_ context.Context, id uint) (*model.Product, error) {
	if p, ok := f.data[id]; ok { // found
		return p, nil
	}
	return nil, repository.ErrNotFound // not found
}

// Update replaces the stored product.
func (f *fakeRepo) Update(_ context.Context, p *model.Product) error { f.data[p.ID] = p; return nil }

// Delete removes the product or returns repository.ErrNotFound.
func (f *fakeRepo) Delete(_ context.Context, id uint) error {
	if _, ok := f.data[id]; !ok { // not found
		return repository.ErrNotFound
	}
	delete(f.data, id) // remove it
	return nil         // success
}

// TestCreateAndGet verifies a product can be created and then fetched.
func TestCreateAndGet(t *testing.T) {
	svc := NewProductService(newFakeRepo()) // service with fake repo
	ctx := context.Background()             // empty context

	// Create a product with surrounding spaces in the name
	created, err := svc.Create(ctx, dto.CreateProductRequest{Name: "  Laptop ", Price: 999.99, ShortDescription: "Fast"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Name should be trimmed by the service
	if created.Name != "Laptop" {
		t.Fatalf("expected trimmed name, got %q", created.Name)
	}
	// Fetch it back by ID
	got, err := svc.GetByID(ctx, created.ID)
	if err != nil || got.Price != 999.99 {
		t.Fatalf("get failed: %v %+v", err, got)
	}
}

// TestGetNotFound verifies a missing product returns a 404 AppError.
func TestGetNotFound(t *testing.T) {
	svc := NewProductService(newFakeRepo())         // service with empty repo
	_, err := svc.GetByID(context.Background(), 42) // ID that does not exist
	appErr, ok := apperror.As(err)                  // extract AppError
	if !ok || appErr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %v", err)
	}
}

// TestUpdatePartial verifies only provided fields are changed.
func TestUpdatePartial(t *testing.T) {
	svc := NewProductService(newFakeRepo()) // service with fake repo
	ctx := context.Background()             // empty context
	created, _ := svc.Create(ctx, dto.CreateProductRequest{Name: "Phone", Price: 100, ShortDescription: "Old"})

	newPrice := 150.0 // only the price changes
	updated, err := svc.Update(ctx, created.ID, dto.UpdateProductRequest{Price: &newPrice})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Price changed, name untouched
	if updated.Price != 150 || updated.Name != "Phone" {
		t.Fatalf("partial update failed: %+v", updated)
	}
}

// TestUpdateEmptyBody verifies an empty update returns 400.
func TestUpdateEmptyBody(t *testing.T) {
	svc := NewProductService(newFakeRepo())                                   // service with fake repo
	_, err := svc.Update(context.Background(), 1, dto.UpdateProductRequest{}) // no fields
	appErr, ok := apperror.As(err)                                            // extract AppError
	if !ok || appErr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %v", err)
	}
}

// TestDelete verifies delete works and a second delete returns 404.
func TestDelete(t *testing.T) {
	svc := NewProductService(newFakeRepo()) // service with fake repo
	ctx := context.Background()             // empty context
	created, _ := svc.Create(ctx, dto.CreateProductRequest{Name: "Mouse", Price: 20})

	if err := svc.Delete(ctx, created.ID); err != nil { // first delete succeeds
		t.Fatalf("unexpected error: %v", err)
	}
	appErr, ok := apperror.As(svc.Delete(ctx, created.ID)) // second delete -> 404
	if !ok || appErr.Code != http.StatusNotFound {
		t.Fatal("expected 404 on second delete")
	}
}
