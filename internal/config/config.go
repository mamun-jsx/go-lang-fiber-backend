// Package config loads application configuration from environment variables.
package config

import (
	"fmt"     // fmt is used to build formatted strings (e.g. the DSN)
	"os"      // os gives access to environment variables
	"strconv" // strconv converts strings to numbers
	"time"    // time is used for duration values

	"github.com/joho/godotenv" // godotenv loads variables from a .env file
)

// Config is the root configuration object for the whole application.
type Config struct {
	App AppConfig // App holds HTTP server / application settings
	DB  DBConfig  // DB holds PostgreSQL connection settings
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name string // Name is the application name (shown in logs / headers)
	Env  string // Env is the running environment: development | production
	Port string // Port is the HTTP port the server listens on
}

// DBConfig holds database connection settings.
type DBConfig struct {
	Host            string        // Host is the PostgreSQL server host
	Port            string        // Port is the PostgreSQL server port
	User            string        // User is the database user name
	Password        string        // Password is the database user password
	Name            string        // Name is the database name
	SSLMode         string        // SSLMode is the PostgreSQL sslmode (disable, require, ...)
	TimeZone        string        // TimeZone is the session time zone
	MaxOpenConns    int           // MaxOpenConns is the max number of open connections in the pool
	MaxIdleConns    int           // MaxIdleConns is the max number of idle connections in the pool
	ConnMaxLifetime time.Duration // ConnMaxLifetime is how long a connection may be reused
}

// DSN builds the PostgreSQL connection string used by GORM.
func (d DBConfig) DSN() string {
	// Return the key=value formatted DSN understood by the pgx driver
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		d.Host, d.User, d.Password, d.Name, d.Port, d.SSLMode, d.TimeZone,
	)
}

// Load reads configuration from the environment (and an optional .env file).
func Load() (*Config, error) {
	// Try to load .env; ignore the error because in production real env vars are used
	_ = godotenv.Load()

	// Build the configuration struct, using sensible defaults where values are missing
	cfg := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "go-lang-fiber-backend"), // application name
			Env:  getEnv("APP_ENV", "development"),            // environment
			Port: getEnv("APP_PORT", "8080"),                  // HTTP port
		},
		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),                                             // DB host
			Port:            getEnv("DB_PORT", "5432"),                                                  // DB port
			User:            getEnv("DB_USER", "postgres"),                                              // DB user
			Password:        getEnv("DB_PASSWORD", ""),                                                  // DB password
			Name:            getEnv("DB_NAME", "product_db"),                                            // DB name
			SSLMode:         getEnv("DB_SSLMODE", "disable"),                                            // SSL mode
			TimeZone:        getEnv("DB_TIMEZONE", "UTC"),                                               // time zone
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),                                         // pool: max open
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),                                         // pool: max idle
			ConnMaxLifetime: time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute, // pool: lifetime
		},
	}

	// Validate that the required values are present
	if cfg.DB.Host == "" || cfg.DB.User == "" || cfg.DB.Name == "" {
		// Return a descriptive error so the app fails fast on bad config
		return nil, fmt.Errorf("config: DB_HOST, DB_USER and DB_NAME are required")
	}

	// Return the fully populated configuration
	return cfg, nil
}

// getEnv returns the value of an env variable or a fallback when it is empty.
func getEnv(key, fallback string) string {
	// Look up the variable in the process environment
	if v, ok := os.LookupEnv(key); ok && v != "" {
		// Variable exists and is not empty, so use it
		return v
	}
	// Otherwise use the provided default value
	return fallback
}

// getEnvInt returns an env variable parsed as int, or a fallback on error.
func getEnvInt(key string, fallback int) int {
	// Read the raw string value
	v := getEnv(key, "")
	// Try converting the string into an integer
	n, err := strconv.Atoi(v)
	// If it is missing or not a valid number, use the default
	if err != nil {
		return fallback
	}
	// Return the parsed number
	return n
}
