package config

import (
	"fmt"
	zfg "github.com/chaindead/zerocfg"
	"github.com/chaindead/zerocfg/env"
)

// Define configuration options
var (
	// Application settings
	configPath = zfg.Str("config.path", "", "path to config file", zfg.Alias("c"))
	appName    = zfg.Str("app.name", "coffee-subscription-service", "application name")
	appEnv     = zfg.Str("app.env", "development", "application environment")
	appPort    = zfg.Uint("app.port", 8080, "application port")
	appDebug   = zfg.Bool("app.debug", true, "debug mode")

	// Database settings
	dbHost     = zfg.Str("db.host", "localhost", "database host")
	dbPort     = zfg.Uint("db.port", 5432, "database port")
	dbName     = zfg.Str("db.name", "coffee_subscriptions", "database name")
	dbUser     = zfg.Str("db.user", "postgres", "database user")
	dbPassword = zfg.Str("db.password", "postgres", "database password", zfg.Secret())
	dbSslMode  = zfg.Str("db.ssl_mode", "disable", "database ssl mode")

	// Stripe settings
	stripeSecretKey     = zfg.Str("stripe.secret_key", "sk_test_your_stripe_secret_key", "stripe secret key", zfg.Secret())
	stripeWebhookSecret = zfg.Str("stripe.webhook_secret", "whsec_your_stripe_webhook_secret", "stripe webhook secret", zfg.Secret())

	// JWT settings
	jwtSecret     = zfg.Str("jwt.secret", "your_jwt_secret_key", "JWT secret key", zfg.Secret())
	jwtExpiration = zfg.Dur("jwt.expiration", 24*60*60, "JWT expiration in seconds")
)

// Config holds all configuration settings
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
	Port  uint
	Debug bool
}

// DBConfig holds database configuration
type DBConfig struct {
	Host       string
	Port       uint
	Name       string
	User       string
	Password   string
	SslMode    string
	DSN        string // Connection string
	MigrateURL string // Connection URL for golang-migrate
}

// StripeConfig holds Stripe API configuration
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

// JWTConfig holds JWT authentication configuration
type JWTConfig struct {
	Secret     string
	Expiration int
}

// Load loads configuration from environment and YAML file
func Load() (*Config, error) {
	// Parse configuration sources
	err := zfg.Parse(
		env.New(), // Environment variables
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	// Create config struct
	cfg := &Config{
		App: AppConfig{
			Name:  *appName,
			Env:   *appEnv,
			Port:  *appPort,
			Debug: *appDebug,
		},
		DB: DBConfig{
			Host:     *dbHost,
			Port:     *dbPort,
			Name:     *dbName,
			User:     *dbUser,
			Password: *dbPassword,
			SslMode:  *dbSslMode,
		},
		Stripe: StripeConfig{
			SecretKey:     *stripeSecretKey,
			WebhookSecret: *stripeWebhookSecret,
		},
		JWT: JWTConfig{
			Secret:     *jwtSecret,
			Expiration: int(*jwtExpiration),
		},
	}

	// Construct database connection string
	cfg.DB.DSN = fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SslMode,
	)

	// Construct URL-formatted connection string for golang-migrate
	cfg.DB.MigrateURL = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SslMode,
	)

	return cfg, nil
}

// PrintConfig prints the current configuration (hiding secrets)
func PrintConfig() {
	fmt.Println("Current configuration:")
	fmt.Println(zfg.Show())
}
