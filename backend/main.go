package main

import (
	"anb-app/src/auth"
	"anb-app/src/database"
	"anb-app/src/storage"
	"anb-app/src/user"
	"anb-app/src/video"
	"anb-app/src/vote"
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: File .env could'nt be found")
	}

	db := database.ConnectDB()

	database.MigrateTables(db)

	jwtSecret := os.Getenv("JWT_SECRET")

	redisAddr := os.Getenv("REDIS_ADDR")

	serverPort := os.Getenv("SERVER_PORT")

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("Could not connect to Redis for caching: %v", err)
	}
	defer redisClient.Close()

	log.Printf("Conectando a Redis en: %s", redisAddr)
	redisOpt := asynq.RedisClientOpt{
		Addr: redisAddr,
	}
	asynqClient := asynq.NewClient(redisOpt)
	defer asynqClient.Close()

	// Auth
	authSvc := auth.NewAuthService(jwtSecret)
	authMiddleware := authSvc.AuthMiddleware()

	// User
	userRepo := user.NewUserRepository(db)
	userSvc := user.NewUserService(userRepo, authSvc)
	userController := user.NewUserController(userSvc)

	// Video - Initialize S3 Storage
	s3Bucket := os.Getenv("S3_BUCKET_NAME")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is required")
	}
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1" // Default region
	}

	storageSvc, err := storage.NewS3StorageService(s3Bucket, awsRegion)
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}
	log.Printf("S3 Storage initialized: bucket=%s, region=%s", s3Bucket, awsRegion)

	videoRepo := video.NewVideoRepository(db)
	videoSvc := video.NewVideoService(videoRepo, asynqClient, redisClient, storageSvc)
	videoController := video.NewVideoController(videoSvc)

	// Vote
	voteRepo := vote.NewVoteRepository(db)
	voteSvc := vote.NewVoteService(voteRepo, db)
	voteController := vote.NewVoteController(voteSvc)

	router := gin.Default()

	// No longer serving static files - videos are served from S3 via presigned URLs

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	apiV1 := router.Group("/api/v1")
	{
		user.SignUpUserRoutes(apiV1, userController)
		video.SignUpVideoRoutes(apiV1, videoController, authMiddleware)
		vote.SignUpVoteRoutes(apiV1, voteController, authMiddleware)
	}

	// Backwards-compatible public endpoint without version: /api/public/videos
	// This directly exposes the same handler as /api/v1/public/videos
	router.GET("/api/public/videos", func(c *gin.Context) {
		videoController.ListPublicVideos(c)
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":   "ok",
			"message":  "ANB API is running",
			"database": "connected",
			"redis":    "connected",
		})
	})

	if err := router.Run(":" + serverPort); err != nil {
		log.Fatalf("Error, server couldn't start: %v", err)
	}
}
