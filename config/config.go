package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Wechat   WechatConfig
	DeepSeek DeepSeekConfig
	MinIO    MinIOConfig
	JWT      JWTConfig
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// RedisConfig holds redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// WechatConfig holds wechat configuration
type WechatConfig struct {
	AppID     string
	AppSecret string
}

// DeepSeekConfig holds deepseek configuration
type DeepSeekConfig struct {
	APIKey string
}

// MinIOConfig holds minio configuration
type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
}

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file if exists
	_ = godotenv.Load()
	log.Println("Environment variables loaded")

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "password"),
			DBName:   getEnv("DB_NAME", "lite_collector"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Wechat: WechatConfig{
			AppID:     getEnv("WX_APP_ID", ""),
			AppSecret: getEnv("WX_APP_SECRET", ""),
		},
		DeepSeek: DeepSeekConfig{
			APIKey: getEnv("DEEPSEEK_API_KEY", ""),
		},
		MinIO: MinIOConfig{
			Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:          getEnvAsBool("MINIO_USE_SSL", false),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "change-me-in-production"),
		},
	}
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var v int
		_, err := fmt.Sscanf(value, "%d", &v)
		if err == nil {
			return v
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		var v bool
		_, err := fmt.Sscanf(value, "%t", &v)
		if err == nil {
			return v
		}
	}
	return defaultValue
}
