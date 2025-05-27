package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort     int
	DatabaseURL    string
	JWTSecret      string
	Environment    string
	CompilerPath   string
	MaxCodeSize    int64
	MaxMemoryUsage int64
}

func LoadConfig() *Config {
	return &Config{
		ServerPort:     getEnvAsInt("SERVER_PORT", 8080),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://localhost:5432/clab?sslmode=disable"),
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		Environment:    getEnv("ENV", "development"),
		CompilerPath:   getEnv("COMPILER_PATH", "/usr/bin/gcc"),
		MaxCodeSize:    getEnvAsInt64("MAX_CODE_SIZE", 1024*1024), // 1MB
		MaxMemoryUsage: getEnvAsInt64("MAX_MEMORY_USAGE", 256*1024*1024), // 256MB
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
} 