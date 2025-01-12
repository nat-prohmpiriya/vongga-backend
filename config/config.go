package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	ServerAddress string

	// MongoDB
	MongoURI string
	MongoDB  string

	// Redis
	RedisURI      string
	RedisPassword string

	// Firebase
	FirebaseCredentialsPath string
	FirebaseStorageBucket   string

	// JWT
	JWTSecret          string
	JWTExpiryHours     int
	RefreshTokenSecret string
	RefreshTokenExpiry int // in days
}

func LoadConfig() *Config {
	// Try to load .env file, but don't error if it doesn't exist
	// This allows us to use environment variables from docker --env-file
	if err := godotenv.Load(); err != nil {
		log.Printf("Note: .env")
	}

	config := &Config{
		ServerAddress:           getEnv("SERVER_ADDRESS", ":8080"),
		MongoURI:                getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:                 getEnv("MONGO_DB", "vongga"),
		RedisURI:                getEnv("REDIS_URI", "localhost:6379"),
		RedisPassword:           getEnv("REDIS_PASSWORD", ""),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		FirebaseStorageBucket:   getEnv("FIREBASE_STORAGE_BUCKET", ""),
		JWTSecret:               getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiryHours:          24,
		RefreshTokenSecret:      getEnv("REFRESH_TOKEN_SECRET", "your-refresh-secret-key"),
		RefreshTokenExpiry:      7,
	}

	// Log loaded configuration (mask sensitive values)
	log.Printf("Loaded configuration:")
	log.Printf("- Server Address: %s", config.ServerAddress)
	log.Printf("- MongoDB URI: %s", maskURI(config.MongoURI))
	log.Printf("- MongoDB Name: %s", config.MongoDB)
	log.Printf("- Redis URI: %s", config.RedisURI)
	log.Printf("- Firebase Credentials Path: %s", config.FirebaseCredentialsPath)
	log.Printf("- Firebase Storage Bucket: %s", config.FirebaseStorageBucket)

	return config
}

// GetJWTExpiry returns JWT expiry duration
func (c *Config) GetJWTExpiry() time.Duration {
	return time.Duration(c.JWTExpiryHours) * time.Hour
}

// GetRefreshTokenExpiry returns refresh token expiry duration
func (c *Config) GetRefreshTokenExpiry() time.Duration {
	return time.Duration(c.RefreshTokenExpiry) * 24 * time.Hour
}

// maskURI masks sensitive information in URIs
func maskURI(uri string) string {
	if uri == "" {
		return ""
	}
	return "***masked***"
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
