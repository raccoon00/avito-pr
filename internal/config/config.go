package config

import (
	"fmt"
	"log"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBProvider string
}

func Load() *Config {
	return &Config{
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "example"),
		DBName:     getEnv("DB_NAME", "prs"),
	}
}

func (c *Config) GetDBConnectionString() string {
	log.Printf("Connecting to %s as %s\n", c.DBName, c.DBUser)
	return fmt.Sprintf(
		"postgres://%s:%s@db:5432/%s",
		c.DBUser, c.DBPassword, c.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
