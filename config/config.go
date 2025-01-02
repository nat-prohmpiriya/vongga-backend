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
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	return &Config{
		// Server
		ServerAddress: getEnv("SERVER_ADDRESS", ":8080"),

		// MongoDB
		MongoURI: getEnv("MONGO_URI", ""),
		MongoDB:  getEnv("MONGO_DB", ""),

		// Redis
		RedisURI:      getEnv("REDIS_URI", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Firebase
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		FirebaseStorageBucket:   getEnv("FIREBASE_STORAGE_BUCKET", ""),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiryHours:     1,
		RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET", ""),
		RefreshTokenExpiry: 30,
	}
}

// GetJWTExpiry returns JWT expiry duration
func (c *Config) GetJWTExpiry() time.Duration {
	return time.Duration(c.JWTExpiryHours) * time.Hour
}

// GetRefreshTokenExpiry returns refresh token expiry duration
func (c *Config) GetRefreshTokenExpiry() time.Duration {
	return time.Duration(c.RefreshTokenExpiry) * 24 * time.Hour
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
