package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
)

// Config carrega configuração a partir de variáveis de ambiente.
type Config struct {
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"app"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"app"`
	DBName     string `env:"DB_NAME" envDefault:"app"`
	DBSSLMode  string `env:"DB_SSLMODE" envDefault:"disable"`

	RedisAddr string `env:"REDIS_ADDR" envDefault:"localhost:6379"`

	JWTSecret        string `env:"JWT_SECRET" envDefault:""`
	JWTExpirationRaw string `env:"JWT_EXPIRATION" envDefault:"24h"`

	AppPort string `env:"APP_PORT" envDefault:"8080"`
	LogLevel string `env:"LOG_LEVEL" envDefault:"info"`

	GormDebug string `env:"GORM_DEBUG" envDefault:"false"`

	ShutdownTimeoutRaw string `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
}

// Load lê e valida a configuração do ambiente.
func Load() (*Config, error) {
	var c Config
	if err := env.Parse(&c); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}
	if c.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return &c, nil
}

// PostgresDSN retorna DSN no formato libpq para GORM/driver postgres.
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

// DatabaseURL retorna URL postgres para golang-migrate CLI.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

// JWTExpiration parseia JWT_EXPIRATION.
func (c *Config) JWTExpiration() (time.Duration, error) {
	d, err := time.ParseDuration(c.JWTExpirationRaw)
	if err != nil {
		return 0, fmt.Errorf("parse JWT_EXPIRATION: %w", err)
	}
	return d, nil
}

// ShutdownTimeout parseia SHUTDOWN_TIMEOUT.
func (c *Config) ShutdownTimeout() (time.Duration, error) {
	d, err := time.ParseDuration(c.ShutdownTimeoutRaw)
	if err != nil {
		return 0, fmt.Errorf("parse SHUTDOWN_TIMEOUT: %w", err)
	}
	return d, nil
}

// GormDebugEnabled indica se o modo debug do GORM deve ser ativado.
func (c *Config) GormDebugEnabled() bool {
	return c.GormDebug == "true" || c.GormDebug == "1"
}
