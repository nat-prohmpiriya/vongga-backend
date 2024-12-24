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
	postRepo := repository.NewPostRepository(db)
	subPostRepo := repository.NewSubPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	followRepo := repository.NewFollowRepository(db)
	friendshipRepo := repository.NewFriendshipRepository(db)

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
	postUseCase := usecase.NewPostUseCase(postRepo, subPostRepo)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, postRepo)
	reactionUseCase := usecase.NewReactionUseCase(reactionRepo, postRepo, commentRepo)
	followUseCase := usecase.NewFollowUseCase(followRepo)
	friendshipUseCase := usecase.NewFriendshipUseCase(friendshipRepo)

	// Initialize Fiber app
	app := fiber.New()

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, HEAD, PUT, PATCH, POST, DELETE",
		ExposeHeaders:    "Content-Length",
		MaxAge:           12 * 3600,
	}))

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check - public endpoint
	app.Get("/api", handler.NewHealthHandler(db, redisClient).Health)

	// Middleware
	app.Use(utils.RequestLogger())

	// Routes
	api := app.Group("/api")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/verifyTokenFirebase", handler.NewAuthHandler(authUseCase).VerifyTokenFirebase)
	auth.Post("/refresh", handler.NewAuthHandler(authUseCase).RefreshToken)
	auth.Post("/logout", handler.NewAuthHandler(authUseCase).Logout)

	// Protected routes (everything under /api except /auth)
	protectedApi := api.Group("")
	protectedApi.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))

	// Create route groups
	users := protectedApi.Group("/users")
	posts := protectedApi.Group("/posts")
	comments := protectedApi.Group("/comments")
	reactions := protectedApi.Group("/reactions")
	follows := protectedApi.Group("/follows")
	friendships := protectedApi.Group("/friendships")

	// Initialize handlers with their respective route groups
	handler.NewUserHandler(users, userUseCase)
	handler.NewPostHandler(posts, postUseCase)
	handler.NewCommentHandler(comments, commentUseCase)
	handler.NewReactionHandler(reactions, reactionUseCase)
	handler.NewFollowHandler(follows, followUseCase)
	handler.NewFriendshipHandler(friendships, friendshipUseCase)
	handler.NewFileHandler(protectedApi, fileRepo)

	// Start server
	log.Fatal(app.Listen(cfg.ServerAddress))
}
