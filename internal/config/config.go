package config

import (
	"time"
)

// Config represents the main application configuration
type Config struct {
	// Application settings
	App AppConfig `yaml:"app" mapstructure:"app"`

	// Server configurations
	Servers map[string]ServerConfig `yaml:"servers" mapstructure:"servers"`

	// Security settings
	Security SecurityConfig `yaml:"security" mapstructure:"security"`

	// Logging configuration
	Logging LoggingConfig `yaml:"logging" mapstructure:"logging"`

	// Monitoring settings
	Monitoring MonitoringConfig `yaml:"monitoring" mapstructure:"monitoring"`
}

// AppConfig contains application-level settings
type AppConfig struct {
	Name        string        `yaml:"name" mapstructure:"name"`
	Version     string        `yaml:"version" mapstructure:"version"`
	Environment string        `yaml:"environment" mapstructure:"environment"`
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxRetries  int           `yaml:"max_retries" mapstructure:"max_retries"`
}

// ServerConfig represents configuration for an MCP server
type ServerConfig struct {
	// Basic connection settings
	Host     string        `yaml:"host" mapstructure:"host"`
	Port     int           `yaml:"port" mapstructure:"port"`
	Protocol string        `yaml:"protocol" mapstructure:"protocol"`
	Timeout  time.Duration `yaml:"timeout" mapstructure:"timeout"`

	// Authentication
	Auth AuthConfig `yaml:"auth" mapstructure:"auth"`

	// TLS settings
	TLS TLSConfig `yaml:"tls" mapstructure:"tls"`

	// Server-specific settings
	Settings map[string]interface{} `yaml:"settings" mapstructure:"settings"`

	// Health check settings
	HealthCheck HealthCheckConfig `yaml:"health_check" mapstructure:"health_check"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	Type     string            `yaml:"type" mapstructure:"type"` // none, basic, token, oauth2
	Username string            `yaml:"username" mapstructure:"username"`
	Password string            `yaml:"password" mapstructure:"password"`
	Token    string            `yaml:"token" mapstructure:"token"`
	Headers  map[string]string `yaml:"headers" mapstructure:"headers"`
}

// TLSConfig contains TLS/SSL settings
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	CertFile   string `yaml:"cert_file" mapstructure:"cert_file"`
	KeyFile    string `yaml:"key_file" mapstructure:"key_file"`
	CAFile     string `yaml:"ca_file" mapstructure:"ca_file"`
	SkipVerify bool   `yaml:"skip_verify" mapstructure:"skip_verify"`
	ServerName string `yaml:"server_name" mapstructure:"server_name"`
}

// HealthCheckConfig contains health check settings
type HealthCheckConfig struct {
	Enabled     bool          `yaml:"enabled" mapstructure:"enabled"`
	Interval    time.Duration `yaml:"interval" mapstructure:"interval"`
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`
	MaxFailures int           `yaml:"max_failures" mapstructure:"max_failures"`
	Endpoint    string        `yaml:"endpoint" mapstructure:"endpoint"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	// JWT settings
	JWT JWTConfig `yaml:"jwt" mapstructure:"jwt"`

	// Rate limiting
	RateLimit RateLimitConfig `yaml:"rate_limit" mapstructure:"rate_limit"`

	// CORS settings
	CORS CORSConfig `yaml:"cors" mapstructure:"cors"`
}

// JWTConfig contains JWT token settings
type JWTConfig struct {
	Secret     string        `yaml:"secret" mapstructure:"secret"`
	Expiration time.Duration `yaml:"expiration" mapstructure:"expiration"`
	Issuer     string        `yaml:"issuer" mapstructure:"issuer"`
}

// RateLimitConfig contains rate limiting settings
type RateLimitConfig struct {
	Enabled  bool          `yaml:"enabled" mapstructure:"enabled"`
	Requests int           `yaml:"requests" mapstructure:"requests"`
	Window   time.Duration `yaml:"window" mapstructure:"window"`
}

// CORSConfig contains CORS settings
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled" mapstructure:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers" mapstructure:"allowed_headers"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `yaml:"level" mapstructure:"level"`
	Format     string `yaml:"format" mapstructure:"format"` // json, text
	Output     string `yaml:"output" mapstructure:"output"` // stdout, stderr, file
	File       string `yaml:"file" mapstructure:"file"`
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`
}

// MonitoringConfig contains monitoring settings
type MonitoringConfig struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Host    string `yaml:"host" mapstructure:"host"`
	Port    int    `yaml:"port" mapstructure:"port"`
	Path    string `yaml:"path" mapstructure:"path"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:        "mcp-cli",
			Version:     "1.0.0",
			Environment: "development",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
		},
		Servers: make(map[string]ServerConfig),
		Security: SecurityConfig{
			JWT: JWTConfig{
				Secret:     "your-secret-key",
				Expiration: 24 * time.Hour,
				Issuer:     "mcp-cli",
			},
			RateLimit: RateLimitConfig{
				Enabled:  true,
				Requests: 100,
				Window:   time.Minute,
			},
			CORS: CORSConfig{
				Enabled:        true,
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
				AllowedHeaders: []string{"*"},
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			Format:     "text",
			Output:     "stdout",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
		},
		Monitoring: MonitoringConfig{
			Enabled: true,
			Host:    "localhost",
			Port:    9090,
			Path:    "/metrics",
		},
	}
}
