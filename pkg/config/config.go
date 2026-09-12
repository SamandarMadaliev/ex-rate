package config

import "os"

type Config struct {
	Environment string

	Server struct {
		Host string
		Port string
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
