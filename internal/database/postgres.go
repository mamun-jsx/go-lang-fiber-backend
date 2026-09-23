// Package database is responsible for creating the PostgreSQL connection via GORM.
package database

import (
	"fmt" // fmt is used to wrap errors with context

	"gorm.io/driver/postgres" // postgres is the GORM driver for PostgreSQL
	"gorm.io/gorm"            // gorm is the ORM library
	"gorm.io/gorm/logger"     // logger controls GORM SQL logging

	"github.com/mamun-jsx/go-lang-fiber-backend/internal/config" // config holds DB settings
	"github.com/mamun-jsx/go-lang-fiber-backend/internal/model"  // model holds DB entities
)

// Connect opens a GORM connection to PostgreSQL and configures the pool.
func Connect(cfg config.DBConfig, env string) (*gorm.DB, error) {
	// Choose SQL log level: verbose in development, only errors otherwise
	logLevel := logger.Error
	// In development we want to see every SQL query
	if env == "development" {
		logLevel = logger.Info
	}

	// Open the connection using the DSN built from the config
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel), // attach the configured logger
	})
	// Return a wrapped error if the connection could not be opened
	if err != nil {
		return nil, fmt.Errorf("database: open connection: %w", err)
	}

	// Get the underlying *sql.DB to configure the connection pool
	sqlDB, err := db.DB()
	// Return an error if we cannot access the raw DB handle
	if err != nil {
		return nil, fmt.Errorf("database: get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)       // limit the number of open connections
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)       // limit the number of idle connections
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime) // recycle connections after this duration

	// Ping the database to make sure the connection is really alive
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	// Return the ready-to-use GORM instance
	return db, nil
}

// Migrate creates / updates database tables for all models.
func Migrate(db *gorm.DB) error {
	// AutoMigrate creates the products table if it does not exist
	if err := db.AutoMigrate(&model.Product{}); err != nil {
		return fmt.Errorf("database: auto migrate: %w", err)
	}
	// Migration finished successfully
	return nil
}

// Close closes the underlying database connection pool.
func Close(db *gorm.DB) error {
	// Get the raw *sql.DB handle
	sqlDB, err := db.DB()
	// Return the error if the handle cannot be retrieved
	if err != nil {
		return err
	}
	// Close all connections in the pool
	return sqlDB.Close()
}
