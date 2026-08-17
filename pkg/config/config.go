package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port              string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
	JWTSecret         string
	TokenTTL          time.Duration
	MigrationsDir     string
	SeedAdminUsername string
	SeedAdminPassword string
}

func Load() Config {
	return Config{
		Port:              getEnv("APP_PORT", "8080"),
		DBHost:            getEnv("DB_HOST", "mysql"),
		DBPort:            getEnv("DB_PORT", "3306"),
		DBUser:            getEnv("DB_USER", "blog"),
		DBPassword:        getEnv("DB_PASSWORD", "blog_password"),
		DBName:            getEnv("DB_NAME", "blog"),
		JWTSecret:         getEnv("JWT_SECRET", "change-me-to-a-long-random-secret"),
		TokenTTL:          time.Duration(getEnvInt("JWT_TTL_HOURS", 24)) * time.Hour,
		MigrationsDir:     getEnv("MIGRATIONS_DIR", "migrations"),
		SeedAdminUsername: getEnv("SEED_ADMIN_USERNAME", "admin"),
		SeedAdminPassword: getEnv("SEED_ADMIN_PASSWORD", "Admin123!"),
	}
}

func (c Config) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci&loc=UTC&multiStatements=true",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
