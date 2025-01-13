package main

import (
	"log"
	"time"

	"context"

	"vongga_api/config"
	"vongga_api/internal/adapter"
	handler "vongga_api/internal/handler/http"
	"vongga_api/internal/handler/websocket"
	"vongga_api/internal/middleware"
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
	db, err := adapter.NewMongoDBClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	redisClient, err := adapter.NewRedisClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	firebaseClient, err := adapter.NewFirebaseProvider(cfg)
	if err != nil {
		log.Fatal(err)
	}
	firebaseAuthClient, err := firebaseClient.Auth(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	tp, err := adapter.TraceProvider(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := tp.Shutdown(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()
	tracer := tp.Tracer("vongga_api")
	defer tracer.Start(context.Background(), "main")

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
		firebaseAuthClient,
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
		AllowOrigins:     "http://localhost:4001, https://dev.dacbok.com, https://dacbok.com", // ต้องระบุ origin ชัดเจนเมื่อใช้ credentials
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Authorization",
		MaxAge:           3600,
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

	handler.NewWebHandler(webGroup)
	// Health check - Ping
	app.Get("/api/ping", handler.NewHealthHandler(db, redisClient).Ping)

	// Middleware
	app.Use(utils.RequestLogger())

	// Routes
	api := app.Group("/api")

	wsHandler := websocket.NewWebSocketHandler()
	wsGroup := api.Group("/ws")
	// wsGroup.Use(middleware.WebsocketAuthMiddleware(authUseCase, tracer))
	wsHandler.RegisterRoutes(wsGroup)

	// Public auth routes
	auth := api.Group("/auth")
	auth.Post("/verifyTokenFirebase", handler.NewAuthHandler(authUseCase, tracer).VerifyTokenFirebase)
	auth.Post("/refresh", handler.NewAuthHandler(authUseCase, tracer).RefreshToken)
	auth.Post("/logout", handler.NewAuthHandler(authUseCase, tracer).Logout)
	auth.Post("/createTestToken", handler.NewAuthHandler(authUseCase, tracer).CreateTestToken)

	// Protected routes
	protectedApi := api.Group("", middleware.HttpAuthMiddleware(authUseCase, tracer))

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
	jaeger := protectedApi.Group("/jaeger")

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
	handler.NewJeagerHandler(jaeger)

	// Start server
	log.Fatal(app.Listen(cfg.ServerAddress))
}
