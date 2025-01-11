package main

import (
	"context"
	"log"
	"time"

	"vongga-api/config"
	_ "vongga-api/docs" // swagger docs
	"vongga-api/internal/auth"
	handler "vongga-api/internal/handler/http"
	"vongga-api/internal/handler/http/middleware"
	"vongga-api/internal/handler/websocket"
	"vongga-api/internal/repository"
	"vongga-api/internal/usecase"
	"vongga-api/utils"

	"vongga-api/infrastucture/firebase"
	"vongga-api/infrastucture/mongodb"
	"vongga-api/infrastucture/redis"

	// เพิ่มตรงนี้
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
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
	firebaseApp, err := firebase.InitFirebase(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create auth adapter
	systemAuthAdapter := auth.NewSystemAuthAdapter(cfg.JWTSecret)

	authClient, err := firebaseApp.Auth(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Initialize MongoDB
	db, err := mongodb.InitMongo(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Redis client
	redisClient, err := redis.InitRedis(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Test Redis connection
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, redisClient)
	postRepo := repository.NewPostRepository(db, redisClient)
	followRepo := repository.NewFollowRepository(db)
	friendshipRepo := repository.NewFriendshipRepository(db)
	notificationRepo := repository.NewNotificationRepository(db, redisClient)
	commentRepo := repository.NewCommentRepository(db, redisClient)
	reactionRepo := repository.NewReactionRepository(db)
	subPostRepo := repository.NewSubPostRepository(db, redisClient)
	storyRepo := repository.NewStoryRepository(db, redisClient)
	chatRepo := repository.NewChatRepository(db)
	fileRepo, err := repository.NewFileStorage(cfg.FirebaseCredentialsPath, cfg.FirebaseStorageBucket)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, userRepo)
	postUseCase := usecase.NewPostUseCase(postRepo, subPostRepo, userRepo, notificationUseCase)
	storyUseCase := usecase.NewStoryUseCase(storyRepo, userRepo)
	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		authClient,
		redisClient,
		cfg.JWTSecret,
		cfg.RefreshTokenSecret,
		cfg.GetJWTExpiry(),
		cfg.GetRefreshTokenExpiry(),
	)
	followUseCase := usecase.NewFollowUseCase(followRepo, notificationUseCase)
	friendshipUseCase := usecase.NewFriendshipUseCase(friendshipRepo, notificationUseCase)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, postRepo, notificationUseCase, userRepo)
	reactionUseCase := usecase.NewReactionUseCase(reactionRepo, postRepo, commentRepo, notificationUseCase)
	subPostUseCase := usecase.NewSubPostUseCase(subPostRepo, postRepo)
	chatUseCase := usecase.NewChatUsecase(chatRepo, userRepo, notificationUseCase)

	// Initialize Fiber app with performance configurations
	app := fiber.New(fiber.Config{
		Prefork:       false,
		ServerHeader:  "Vongga",
		StrictRouting: true,
		CaseSensitive: true,
		BodyLimit:     4 * 1024 * 1024, // 4MB
		Concurrency:   256,
	})

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		MaxAge:       3600,
	}))

	// Add compression middleware
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Add cache middleware
	app.Use(cache.New(cache.Config{
		Expiration:   30 * time.Minute,
		CacheControl: true,
	}))

	// Swagger
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Health check - public endpoint
	app.Get("/api", handler.NewHealthHandler(db, redisClient).Health)
	// Health check - Ping
	app.Get("/api/ping", handler.NewHealthHandler(db, redisClient).Ping)

	// Middleware
	app.Use(utils.RequestLogger())

	// Routes
	api := app.Group("/api")

	// WebSocket endpoint (outside protected routes)
	websocket.NewWebSocketHandler(api, chatUseCase, systemAuthAdapter)

	// Public auth routes
	auth := api.Group("/auth")
	auth.Post("/verifyTokenFirebase", handler.NewAuthHandler(authUseCase).VerifyTokenFirebase)
	auth.Post("/refresh", handler.NewAuthHandler(authUseCase).RefreshToken)
	auth.Post("/logout", handler.NewAuthHandler(authUseCase).Logout)
	auth.Post("/createTestToken", handler.NewAuthHandler(authUseCase).CreateTestToken)

	// Protected routes
	protectedApi := api.Group("", middleware.AuthMiddleware(cfg.JWTSecret))

	// Create route groups
	users := protectedApi.Group("/users")
	posts := protectedApi.Group("/posts")
	comments := protectedApi.Group("/comments")
	reactions := protectedApi.Group("/reactions")
	follows := protectedApi.Group("/follows")
	friendships := protectedApi.Group("/friendships")
	notifications := protectedApi.Group("/notifications")
	stories := protectedApi.Group("/stories")
	chats := protectedApi.Group("/chat")

	// Initialize handlers with their respective route groups
	handler.NewUserHandler(users, userUseCase)
	handler.NewFollowHandler(follows, followUseCase)
	handler.NewFriendshipHandler(friendships, friendshipUseCase)
	handler.NewPostHandler(posts, postUseCase)
	handler.NewSubPostHandler(posts, subPostUseCase)
	handler.NewCommentHandler(comments, commentUseCase, userUseCase)
	handler.NewReactionHandler(reactions, reactionUseCase)
	handler.NewNotificationHandler(notifications, notificationUseCase)
	handler.NewStoryHandler(stories, storyUseCase)
	handler.NewFileHandler(protectedApi, fileRepo)
	handler.NewChatHandler(chats, chatUseCase)

	// Start server
	log.Fatal(app.Listen(cfg.ServerAddress))
}
