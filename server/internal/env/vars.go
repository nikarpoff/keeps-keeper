// Package env loads the service configuration from environment variables.
package env

import (
	"errors"

	cenv "github.com/caarlos0/env/v6"
)

// Vars holds the service configuration.
type Vars struct {
	// server variables
	AppServerPort            int    `env:"APP_SERVER_PORT" envDefault:"80"`
	AppServerShutdownTimeout int    `env:"APP_SERVER_SHUTDOWN_TIMEOUT" envDefault:"10"`
	AppStaticDir             string `env:"APP_STATIC_DIR" envDefault:"./static"`

	// auth variables
	AppJWTSecret    string `env:"APP_JWT_SECRET"`
	AppJWTTTLHours  int    `env:"APP_JWT_TTL_HOURS" envDefault:"720"`
	AppCookieSecure bool   `env:"APP_COOKIE_SECURE" envDefault:"true"`

	// postgres variables
	PostgresDSN string `env:"POSTGRES_DSN"`
}

// Load parses the environment and checks required variables.
func Load() (*Vars, error) {
	var v Vars
	if err := cenv.Parse(&v); err != nil {
		return nil, err
	}
	if v.AppJWTSecret == "" {
		return nil, errors.New("APP_JWT_SECRET environment variable is required")
	}
	if v.PostgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN environment variable is required")
	}
	return &v, nil
}
