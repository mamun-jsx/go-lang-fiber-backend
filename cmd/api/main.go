// Command api is the entry point of the Product CRUD REST API.
// It wires all layers together (config -> DB -> repository -> service -> handler -> router)
// and starts the HTTP server with graceful shutdown.
package main

import (
	"log"       // log prints startup / shutdown messages
	"os"        // os is used for OS signals
	"os/signal" // signal listens for Ctrl+C / SIGTERM
	"syscall"   // syscall provides the SIGTERM constant
	"time"      // time is used for the shutdown timeout

	"github.com/gofiber/fiber/v3" // fiber is the web framework

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/config"     // configuration loader
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/database"   // DB connection + migration
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/handler"    // HTTP handlers
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/middleware" // global error handler
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/repository" // data access layer
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/router"     // route registration
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/service"    // business logic layer
	"github.com/mamun-jsx/go-lang-fiber-backend/pkg/validator"       // request validator
)

func main() {
	// ---------- 1. Load configuration ----------
	cfg, err := config.Load()
	// Stop immediately if the configuration is invalid
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ---------- 2. Connect to PostgreSQL ----------
	db, err := database.Connect(cfg.DB, cfg.App.Env)
	// Stop if the database is unreachable
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	// Make sure the connection pool is closed when main returns
	defer func() {
		// Close the DB and log any error
		if err := database.Close(db); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()
	// Log successful connection
	log.Println("database connected")

	// ---------- 3. Run migrations ----------
	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	// Log successful migration
	log.Println("database migrated")

	// ---------- 4. Build dependencies (manual dependency injection) ----------
	productRepo := repository.NewProductRepository(db)                           // repository needs the DB
	productService := service.NewProductService(productRepo)                     // service needs the repository
	productHandler := handler.NewProductHandler(productService, validator.New()) // handler needs the service

	// ---------- 5. Create the Fiber app ----------
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,            // name shown in the startup banner
		ErrorHandler: middleware.ErrorHandler, // central error -> JSON conversion
		ReadTimeout:  10 * time.Second,        // max time to read a request
		WriteTimeout: 10 * time.Second,        // max time to write a response
		IdleTimeout:  60 * time.Second,        // max keep-alive idle time
	})

	// ---------- 6. Register routes ----------
	router.Setup(app, productHandler)

	// ---------- 7. Start server in a goroutine ----------
	go func() {
		// Listen on all interfaces at the configured port
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	// ---------- 8. Graceful shutdown ----------
	quit := make(chan os.Signal, 1)                    // channel to receive OS signals
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM) // listen for Ctrl+C and SIGTERM
	<-quit                                             // block until a signal arrives
	log.Println("shutting down server...")             // inform that shutdown started

	// Give in-flight requests up to 10 seconds to finish
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}
	// Final log line before the deferred DB close runs
	log.Println("server exited")
}
