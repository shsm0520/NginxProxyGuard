package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseURL     string
	RedisURL        string
	JWTSecret       string
	Environment     string
	NginxConfigPath string
	NginxCertsPath  string
	NginxContainer  string
	ACMEEmail       string
	ACMEStaging     bool
	LogCollection   bool
	BackupPath      string
}

func Load() *Config {
	// Load .env file if exists
	godotenv.Load()

	return &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@db:5432/nginx_proxy_guard?sslmode=disable"),
		RedisURL:        getEnv("REDIS_URL", ""),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		NginxConfigPath: getEnv("NGINX_CONFIG_PATH", "/etc/nginx/conf.d"),
		NginxCertsPath:  getEnv("NGINX_CERTS_PATH", "/etc/nginx/certs"),
		NginxContainer:  getEnv("NGINX_CONTAINER", "npg-proxy"),
		ACMEEmail:       getEnv("ACME_EMAIL", ""),
		ACMEStaging:     getEnv("ACME_STAGING", "true") == "true",
		LogCollection:   getEnv("LOG_COLLECTION", "true") == "true",
		BackupPath:      getEnv("BACKUP_PATH", "/data/backups"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
