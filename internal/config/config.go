package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	JwtSecret      string
	DBURL          string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOUseSSL    bool
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:           getEnv("PORT", "8085"),
		JwtSecret:      getEnv("JWT_SECRET", ""),
		DBURL:          getEnv("DB_URL", ""),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", ""),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", ""),
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
	}

	if cfg.DBURL == "" {
		log.Println("WARNING: DB_URL is empty")
	}
	if cfg.JwtSecret == "" {
		log.Println("WARNING: JWT_SECRET is empty")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
