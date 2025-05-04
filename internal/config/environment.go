package config

import (
	"os"
	"strconv"
	"strings"
)

// IsDevelopment checks if the environment is development
func (c *Config) IsDevelopment() bool {
	return strings.ToLower(c.App.Env) == "development"
}

// IsProduction checks if the environment is production
func (c *Config) IsProduction() bool {
	return strings.ToLower(c.App.Env) == "production"
}

// IsTest checks if the environment is test
func (c *Config) IsTest() bool {
	return strings.ToLower(c.App.Env) == "test"
}

// GetEnv gets an environment variable or returns fallback value
func GetEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GetEnvBool gets a boolean environment variable or returns fallback value
func GetEnvBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fallback
		}
		return b
	}
	return fallback
}

// GetEnvInt gets an integer environment variable or returns fallback value
func GetEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		i, err := strconv.Atoi(value)
		if err != nil {
			return fallback
		}
		return i
	}
	return fallback
}
