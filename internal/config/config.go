package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	App      AppConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	SSLMode     string
	DatabaseURL string // For Cloud providers like neon
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string
	Port string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret      string
	ExpiryHours time.Duration
}

// AppConfig holds general application configuration
type AppConfig struct {
	Environment string
	LogLevel    string
}

// Load func loads configuration from enviroment variables
func Load() (*Config, error) {
	// Load .env file if it exists (developemt)
	if err := godotenv.Load(); err != nil {
		// It's okay if .end esn't exist in production
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:        getEnv("DB_HOST", "localhost"),
			Port:        getEnv("DB_PORT", "5432"),
			User:        getEnv("DB_USER", "postgres"),
			Password:    getEnv("DB_PASSWORD", ""),
			Database:    getEnv("DB_NAME", "todoapp"),
			SSLMode:     getEnv("DB_SSL_MODE", "require"), // Default to 'require' for cloud DBs
			DatabaseURL: getEnv("DATABASE_URL", ""),       // Neon connection string
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", ""),
			ExpiryHours: time.Duration(getEnvAsInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
		},
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// If DATABASE_URL is provided, skip individual field validation
	if c.Database.DatabaseURL != "" {
		// Just validate JWT Secret
		if c.JWT.Secret == "" {
			return fmt.Errorf("JWT secret is required")
		}
		if len(c.JWT.Secret) < 32 {
			return fmt.Errorf("JWT secret must be at least 32 characters long")
		}
		return nil
	}
	// Otherwise validate individual database fields
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}

	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if len(c.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret invalid")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variables as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
