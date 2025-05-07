// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	App    AppConfig
	DB     DBConfig
	Stripe StripeConfig
	JWT    JWTConfig
}

// AppConfig holds application-specific configuration
type AppConfig struct {
	Name  string
	Env   string
	Port  int
	Debug bool
}

// DBConfig holds database configuration
type DBConfig struct {
	Host       string
	Port       int
	Name       string
	User       string
	Password   string
	SSLMode    string
	DSN        string
	MigrateURL string
}

// StripeConfig holds Stripe API configuration
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret     string
	Expiration string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	cfg := &Config{
		App: AppConfig{
			Name:  getEnv("APP_NAME", "coffee-subscription-service"),
			Env:   getEnv("APP_ENV", "development"),
			Port:  getEnvAsInt("APP_PORT", 8080),
			Debug: getEnvAsBool("APP_DEBUG", true),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Name:     getEnv("DB_NAME", "coffee_subscriptions"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Stripe: StripeConfig{
			SecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
			WebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your_jwt_secret_key"),
			Expiration: getEnv("JWT_EXPIRATION", "24h"),
		},
	}

	// Validate required configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Construct database connection string
	cfg.DB.DSN = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode,
	)

	// Construct URL-formatted connection string for golang-migrate
	cfg.DB.MigrateURL = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode,
	)

	return cfg, nil
}

// validate checks if all required configuration is present
func (c *Config) validate() error {
	// In production, some values are required
	if c.App.Env == "production" {
		if c.Stripe.SecretKey == "" {
			return fmt.Errorf("STRIPE_SECRET_KEY is required in production")
		}
		if c.Stripe.WebhookSecret == "" {
			return fmt.Errorf("STRIPE_WEBHOOK_SECRET is required in production")
		}
		if c.JWT.Secret == "your_jwt_secret_key" {
			return fmt.Errorf("JWT_SECRET must be changed in production")
		}
	}

	return nil
}

// Helper functions to get environment variables with default values
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
