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
	FirebaseBucketName     string

	// JWT
	JWTSecret            string
	JWTExpiryHours      int
	RefreshTokenSecret   string
	RefreshTokenExpiry   int // in days
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
		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:  getEnv("MONGO_DB", "vongga"),

		// Redis
		RedisURI:      getEnv("REDIS_URI", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// Firebase
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "firebase-credentials.json"),
		FirebaseBucketName:     getEnv("FIREBASE_BUCKET_NAME", "vongga-platform.appspot.com"),

		// JWT
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiryHours:    24,
		RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET", "your-refresh-secret-key"),
		RefreshTokenExpiry: 7,
	}
}

func (c *Config) GetJWTExpiry() time.Duration {
	return time.Duration(c.JWTExpiryHours) * time.Hour
}

func (c *Config) GetRefreshTokenExpiry() time.Duration {
	return time.Duration(c.RefreshTokenExpiry) * 24 * time.Hour
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
