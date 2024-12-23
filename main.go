package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/config"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/delivery/http/handler"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/delivery/http/middleware"
	_ "github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/docs" // swagger docs
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/repository"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/usecase"
	"github.com/prohmpiriya_phonumnuaisuk/vongga-platform/vongga-backend/utils"
	"github.com/redis/go-redis/v9"
)

// @title Vongga Backend API
// @version 1.0
// @description This is the Vongga backend server API documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email your.email@vongga.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize Firebase Admin
	firebaseApp, err := config.InitFirebase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Initialize MongoDB
	db, err := config.InitMongo(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisURI,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Test Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	fileRepo, err := repository.NewFileStorage("firebase-credentials.json", cfg.FirebaseStorageBucket)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		authClient,
		redisClient,
		cfg.JWTSecret,
		cfg.RefreshTokenSecret,
		cfg.GetJWTExpiry(),
		cfg.GetRefreshTokenExpiry(),
	)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userUseCase)
	authHandler := handler.NewAuthHandler(authUseCase)
	healthHandler := handler.NewHealthHandler(db, redisClient)
	fileHandler := handler.NewFileHandler(fileRepo)

	// Initialize Fiber app
	app := fiber.New()

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check - public endpoint
	app.Get("/api/health", healthHandler.Health)

	// Middleware
	app.Use(utils.RequestLogger())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// Routes
	api := app.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/verifyTokenFirebase", authHandler.VerifyTokenFirebase)
	auth.Post("/refresh", authHandler.RefreshToken)
	auth.Post("/logout", authHandler.Logout)

	// Protected routes (everything under /api except /auth)
	protectedApi := api.Group("")
	protectedApi.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))

	// User routes
	users := protectedApi.Group("/users")
	users.Post("/", userHandler.CreateOrUpdateUser)
	users.Get("/profile", userHandler.GetProfile)
	users.Patch("/", userHandler.UpdateUser)
	users.Get("/check-username", userHandler.CheckUsername)

	// File upload route
	protectedApi.Post("/upload", fileHandler.Upload)

	// Start server
	log.Fatal(app.Listen(cfg.ServerAddress))
}
