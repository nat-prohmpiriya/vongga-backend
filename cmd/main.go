package main

import (
	"log"
	"time"

	"vongga_api/config"
	"vongga_api/internal/adapter/mongodb"
	"vongga_api/internal/adapter/otel"
	"vongga_api/internal/adapter/redis"
	handler "vongga_api/internal/handler/http"
	"vongga_api/internal/handler/websocket"
	"vongga_api/internal/repository"
	"vongga_api/internal/usecase"
	"vongga_api/utils"

	// เพิ่มตรงนี้
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect External Service
	db, err := mongodb.NewMongoDBClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	redisClient, err := redis.NewRedisClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	authClient, err := firebase.NewFirebaseClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	tracer, systemAuthAdapter := otel.NewTracer(cfg)

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
		tracer,
	)
	followUseCase := usecase.NewFollowUseCase(followRepo, notificationUseCase, tracer)
	friendshipUseCase := usecase.NewFriendshipUseCase(friendshipRepo, notificationUseCase, tracer)
	commentUseCase := usecase.NewCommentUseCase(commentRepo, postRepo, notificationUseCase, userRepo, tracer)
	reactionUseCase := usecase.NewReactionUseCase(reactionRepo, postRepo, commentRepo, notificationUseCase, tracer)
	subPostUseCase := usecase.NewSubPostUseCase(subPostRepo, postRepo, tracer)
	chatUseCase := usecase.NewChatUsecase(chatRepo, userRepo, notificationUseCase, tracer)

	// Initialize Fiber app with performance configurations
	app := fiber.New(fiber.Config{
		Prefork:       false,           // ไม่ใช้ prefork mode (ไม่แยก process)
		ServerHeader:  "Vongga",        // ชื่อ server ใน header
		StrictRouting: true,            // route ต้องตรงแบบ strict (/foo != /foo/)
		CaseSensitive: true,            // route เป็น case sensitive (/Foo != /foo)
		BodyLimit:     4 * 1024 * 1024, // จำกัดขนาด request body 4MB
		Concurrency:   256,             // จำนวน concurrent connections สูงสุด
	})

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",                                           // อนุญาตทุก origin
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",                 // HTTP methods ที่อนุญาต
		AllowHeaders: "Origin, Content-Type, Accept, Authorization", // headers ที่อนุญาต
		MaxAge:       3600,                                          // browser จะ cache CORS response นานเท่าไร
	}))

	// Add compression middleware
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed, // ระดับการบีบอัดข้อมูล
	}))

	// Add cache middleware
	app.Use(cache.New(cache.Config{
		Expiration:   30 * time.Minute, // cache หมดอายุใน 30 นาที
		CacheControl: true,             // ใช้ Cache-Control header
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
	websocket.NewWebSocketHandler(api, chatUseCase, systemAuthAdapter, tracer)

	// Public auth routes
	auth := api.Group("/auth")
	auth.Post("/verifyTokenFirebase", handler.NewAuthHandler(authUseCase, tracer).VerifyTokenFirebase)
	auth.Post("/refresh", handler.NewAuthHandler(authUseCase, tracer).RefreshToken)
	auth.Post("/logout", handler.NewAuthHandler(authUseCase, tracer).Logout)
	auth.Post("/createTestToken", handler.NewAuthHandler(authUseCase, tracer).CreateTestToken)

	// Protected routes
	protectedApi := api.Group("", middleware.AuthMiddleware(cfg.JWTSecret, tracer))

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
