// Package router registers all HTTP routes and global middleware.
package router

import (
	"github.com/gofiber/fiber/v3"                      // fiber is the web framework
	"github.com/gofiber/fiber/v3/middleware/cors"      // cors allows cross-origin requests
	"github.com/gofiber/fiber/v3/middleware/logger"    // logger logs every request
	"github.com/gofiber/fiber/v3/middleware/recover"   // recover turns panics into 500 errors
	"github.com/gofiber/fiber/v3/middleware/requestid" // requestid adds an X-Request-ID header

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/handler" // HTTP handlers
	"github.com/mamun-jsx/go-lang-fiber-backend/pkg/response"     // JSON envelope helpers
)

// Setup attaches global middleware and all API routes to the Fiber app.
func Setup(app *fiber.App, productHandler *handler.ProductHandler) {
	app.Use(recover.New())   // catch panics so one bad request cannot crash the server
	app.Use(requestid.New()) // give each request a unique ID for tracing
	app.Use(logger.New())    // log method, path, status and latency of each request
	app.Use(cors.New())      // enable CORS with default (permissive) settings

	// Health check endpoint used by load balancers / monitoring
	app.Get("/health", func(c fiber.Ctx) error {
		// Always respond 200 OK while the process is alive
		return response.Success(c, fiber.StatusOK, "service is healthy", nil)
	})

	// Group all versioned API routes under /api/v1
	api := app.Group("/api/v1")

	// Group all product routes under /api/v1/products
	products := api.Group("/products")
	products.Post("/", productHandler.Create)      // POST   /api/v1/products      -> create
	products.Get("/", productHandler.List)         // GET    /api/v1/products      -> list
	products.Get("/:id", productHandler.GetByID)   // GET    /api/v1/products/:id  -> read one
	products.Put("/:id", productHandler.Update)    // PUT    /api/v1/products/:id  -> update
	products.Delete("/:id", productHandler.Delete) // DELETE /api/v1/products/:id  -> delete
}
