package config

import (
	"os"
	"time"
)

type Config struct {
	Environment string

	Server struct {
		Host              string
		Port              string
		ReadTimeout       time.Duration
		WriteTimeout      time.Duration
		IdleTimeout       time.Duration
		ReadHeaderTimeout time.Duration
		ShutdownTimeout   time.Duration
	}

	DB struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SSLMode  string
	}
}

func NewConfig() (*Config, error) {
	c := &Config{}

	c.Environment = getEnv("APP_ENV", "local")

	c.Server.Host = getEnv("HTTP_HOST", "0.0.0.0")
	c.Server.Port = getEnv("HTTP_PORT", "9090")
	c.Server.ReadTimeout = getEnvDuration("HTTP_READ_TIMEOUT", 5*time.Second)
	c.Server.WriteTimeout = getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	c.Server.IdleTimeout = getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	c.Server.ReadHeaderTimeout = getEnvDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	c.Server.ShutdownTimeout = getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)

	c.DB.Host = getEnv("DATABASE_HOST", "localhost")
	c.DB.Port = getEnv("DATABASE_PORT", "5432")
	c.DB.User = getEnv("DATABASE_USER", "postgres")
	c.DB.Password = getEnv("DATABASE_PASSWORD", "")
	c.DB.Name = getEnv("DATABASE_NAME", "postgres")
	c.DB.SSLMode = getEnv("DATABASE_SSLMODE", "disable")

	return c, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
