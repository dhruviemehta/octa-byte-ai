package config

import (
    "os"
    "strconv"
)

type Config struct {
    Port     string
    LogLevel string
    Database DatabaseConfig
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Name     string
    User     string
    Password string
    SSLMode  string
}

func Load() *Config {
    return &Config{
        Port:     getEnv("PORT", "8080"),
        LogLevel: getEnv("LOG_LEVEL", "info"),
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnvAsInt("DB_PORT", 5432),
            Name:     getEnv("DB_NAME", "appdb"),
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
            SSLMode:  getEnv("DB_SSL_MODE", "require"), // require for AWS, disable for local
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
