
// internal/config/environment.go
package config

// Environment constants
const (
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"
	EnvTest        = "test"
)

// IsDevelopment checks if the current environment is development
func (c *Config) IsDevelopment() bool {
	return c.App.Env == EnvDevelopment
}

// IsProduction checks if the current environment is production
func (c *Config) IsProduction() bool {
	return c.App.Env == EnvProduction
}

// IsTest checks if the current environment is test
func (c *Config) IsTest() bool {
	return c.App.Env == EnvTest
}