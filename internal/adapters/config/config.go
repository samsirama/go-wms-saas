package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Security SecurityConfig
}

type AppConfig struct {
	Env  string
	Name string
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxConnections  int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type SecurityConfig struct {
	JWTSecret      string
	JWTExpiry      time.Duration
	AllowedOrigins string
}

func LoadConfig() *Config {
	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Name: getEnv("APP_NAME", "WMS SaaS"),
		},
		Server: ServerConfig{
			Port:         getEnv("API_PORT", "8080"),
			ReadTimeout:  time.Duration(getEnvAsInt("SERVER_READ_TIMEOUT", 10)) * time.Second,
			WriteTimeout: time.Duration(getEnvAsInt("SERVER_WRITE_TIMEOUT", 10)) * time.Second,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "wmsadmin"),
			Password:        getEnv("DB_PASSWORD", "wmspassword"),
			DBName:          getEnv("DB_NAME", "wms_saas"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
			ConnMaxLifetime: time.Duration(getEnvAsInt("DB_MAX_LIFETIME_MINUTES", 30)) * time.Minute,
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Security: SecurityConfig{
			JWTSecret:      getEnv("JWT_SECRET", "change-me-in-production"),
			JWTExpiry:      time.Duration(getEnvAsInt("JWT_EXPIRY_HOURS", 24)) * time.Hour,
			AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
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
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
