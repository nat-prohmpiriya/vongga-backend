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
	"vongga-api/infrastucture/otel"
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

	tracerProvider, err := otel.TraceProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	tracer := tracerProvider.Tracer("vongga-api")

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
	userRepo := repository.NewUserRepository(db, redisClient, tracer)
	postRepo := repository.NewPostRepository(db, redisClient, tracer)
	followRepo := repository.NewFollowRepository(db, tracer)
	friendshipRepo := repository.NewFriendshipRepository(db, tracer)
	notificationRepo := repository.NewNotificationRepository(db, redisClient, tracer)
	commentRepo := repository.NewCommentRepository(db, redisClient, tracer)
	reactionRepo := repository.NewReactionRepository(db, tracer)
	subPostRepo := repository.NewSubPostRepository(db, redisClient, tracer)
	storyRepo := repository.NewStoryRepository(db, redisClient, tracer)
	chatRepo := repository.NewChatRepository(db, tracer)
	fileRepo, err := repository.NewFileStorage(cfg.FirebaseCredentialsPath, cfg.FirebaseStorageBucket, tracer)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(userRepo, tracer)
	notificationUseCase := usecase.NewNotificationUseCase(notificationRepo, userRepo, tracer)
	postUseCase := usecase.NewPostUseCase(postRepo, subPostRepo, userRepo, notificationUseCase, tracer)
	storyUseCase := usecase.NewStoryUseCase(storyRepo, userRepo, tracer)
	authUseCase := usecase.NewAuthUseCase(
		userRepo,
		authClient,
		redisClient,
		cfg.JWTSecret,
		cfg.RefreshTokenSecret,
		cfg.FindJWTExpiry(),
		cfg.FindRefreshTokenExpiry(),
	)
	followUseCase := usecase.NewFollowUseCase(followRepo, notificationUseCase, tracer)
	friendshipUseCase := usecase.NewFriendshipUseCase(friendshipRepo, notificationUseCase, tracer)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, postRepo, notificationUseCase, userRepo, tracer)
	reactionUseCase := usecase.NewReactionUseCase(reactionRepo, postRepo, commentRepo, notificationUseCase, tracer)
	subPostUseCase := usecase.NewSubPostUseCase(subPostRepo, postRepo, tracer)
	chatUseCase := usecase.NewChatUsecase(chatRepo, userRepo, notificationUseCase, tracer)

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

	app.Static("/", "/app")
	// Health check - public endpoint
	app.Get("/api", handler.NewHealthHandler(db, redisClient).Health)
	webGroup := app.Group("api/web")
	jaeger := app.Group("/jaeger")
	handler.NewJeagerHandler(jaeger)
	handler.NewWebHandler(webGroup)
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
	auth.Post("/verifyTokenFirebase", handler.NewAuthHandler(authUseCase).VerifyTokenFirebase, tracer)
	auth.Post("/refresh", handler.NewAuthHandler(authUseCase).RefreshToken, tracer)
	auth.Post("/logout", handler.NewAuthHandler(authUseCase).Logout, tracer)
	auth.Post("/createTestToken", handler.NewAuthHandler(authUseCase).CreateTestToken, tracer)

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
	handler.NewUserHandler(users, userUseCase, tracer)
	handler.NewFollowHandler(follows, followUseCase, tracer)
	handler.NewFriendshipHandler(friendships, friendshipUseCase, tracer)
	handler.NewPostHandler(posts, postUseCase, tracer)
	handler.NewSubPostHandler(posts, subPostUseCase, tracer)
	handler.NewCommentHandler(comments, commentUseCase, userUseCase, tracer)
	handler.NewReactionHandler(reactions, reactionUseCase, tracer)
	handler.NewNotificationHandler(notifications, notificationUseCase, tracer)
	handler.NewStoryHandler(stories, storyUseCase, tracer)
	handler.NewFileHandler(protectedApi, fileRepo, tracer)
	handler.NewChatHandler(chats, chatUseCase, tracer)

	// Start server
	log.Fatal(app.Listen(cfg.ServerAddress))
}
