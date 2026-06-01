package config

import (
	"fmt"
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
	_ = godotenv.Load("../.env")

	cfg := &Config{
		Port:           getEnv("CHAT_SERVICE_PORT", getEnv("PORT", "8085")),
		JwtSecret:      getEnv("JWT_SECRET", ""),
		DBURL:          getEnv("DB_URL", defaultDBURL()),
		MinIOEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getEnv("MINIO_ACCESS_KEY", getEnv("MINIO_ROOT_USER", "minioadmin")),
		MinIOSecretKey: getEnv("MINIO_SECRET_KEY", getEnv("MINIO_ROOT_PASSWORD", "minioadmin")),
		MinIOUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
	}

	if cfg.JwtSecret == "" {
		log.Println("WARNING: JWT_SECRET is empty — WebSocket auth will fail")
	}

	return cfg
}

func defaultDBURL() string {
	user := getEnv("DB_USERNAME", getEnv("POSTGRES_USER", "postgres"))
	pass := getEnv("DB_PASSWORD", getEnv("POSTGRES_PASSWORD", "postgres"))
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("CHAT_DB_PORT", "5441")
	name := getEnv("CHAT_DB_NAME", "chat_db")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
