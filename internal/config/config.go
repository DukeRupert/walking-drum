package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Stripe   StripeConfig
	JWT      JWTConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name  string
	Env   string
	Port  int
	Debug bool
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

// StripeConfig holds Stripe API configuration
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret    string
	ExpiresIn time.Duration
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Parse application configuration
	appConfig := AppConfig{
		Name:  getEnv("APP_NAME", "coffee-subscription-service"),
		Env:   getEnv("APP_ENV", "development"),
		Port:  getEnvAsInt("APP_PORT", 8080),
		Debug: getEnvAsBool("APP_DEBUG", true),
	}

	// Parse database configuration
	dbConfig := DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		Name:     getEnv("DB_NAME", "coffee_subscriptions"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	// Parse Stripe configuration
	stripeConfig := StripeConfig{
		SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
	}

	// Parse JWT configuration
	jwtExpiration, err := time.ParseDuration(getEnv("JWT_EXPIRATION", "24h"))
	if err != nil {
		jwtExpiration = 24 * time.Hour
	}

	jwtConfig := JWTConfig{
		Secret:    getEnv("JWT_SECRET", "your_jwt_secret_key"),
		ExpiresIn: jwtExpiration,
	}

	// Return the complete configuration
	return &Config{
		App:      appConfig,
		Database: dbConfig,
		Stripe:   stripeConfig,
		JWT:      jwtConfig,
	}, nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
