package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for our application
type Config struct {
	// Database configuration
	Database DatabaseConfig
	// Server configuration
	Server ServerConfig
	// Environment
	Environment string

	// Redis configuration
	Redis RedisConfig

	// Auth configuration
	Auth AuthConfig

	ProjectName string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret      string
	JWTExpiration  time.Duration
	SessionExpTime time.Duration
}

// Load reads configuration from environment variables
func Load() *Config {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	return &Config{
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "127.0.0.1"),
			Port:            getEnvAsInt("DB_PORT", 3306),
			User:            getEnv("DB_USER", "user"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "tests"),
			URL:             getEnv("DATABASE_URL", "mysql://root:@tcp(127.0.0.1:3306)/tests?parseTime=true"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_CONN_MAX_LIFETIME", 3600)) * time.Second,
		},
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 5)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 10)) * time.Second,
			IdleTimeout:  time.Duration(getEnvAsInt("SERVER_IDLE_TIMEOUT", 30)) * time.Second,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "127.0.0.1"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("JWT_SECRET", "SECRET"),
			JWTExpiration:  time.Duration(getEnvAsInt("JWT_EXPIRATION", 86400)) * time.Second,
			SessionExpTime: time.Duration(getEnvAsInt("SESSION_EXPIRATION", 86400)) * time.Second,
		},
		Environment: getEnv("ENV", "development"),
		ProjectName: getEnv("PROJECT_NAME", "project-name-test"),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using fallback: %d", key, value, fallback)
	}
	return fallback
}

// GetDSN returns database connection string for Go applications
func (c *Config) GetDSN() string {
	return c.Database.User + ":" + c.Database.Password + "@tcp(" + c.Database.Host + ":" + strconv.Itoa(c.Database.Port) + ")/" + c.Database.Name + "?parseTime=true"
}

// GetDatabaseURL returns DATABASE_URL for dbmate
func (c *Config) GetDatabaseURL() string {
	return c.Database.URL
}

// IsDevelopment checks if environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
