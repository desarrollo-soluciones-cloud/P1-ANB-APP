package main

import (
	"anb-app/src/auth"
	"anb-app/src/user"
	"anb-app/src/video"
	"anb-app/src/vote"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("anb.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&user.User{}, &video.Video{}, &vote.Vote{})

	jwtSecret := "MI_CLAVE_SECRETA_SUPREMAMENTE_SEGURA"

	// --- 2. CREAMOS EL CLIENTE DE ASYNQ ---
	redisOpt := asynq.RedisClientOpt{
		Addr: "localhost:6379", // La dirección de nuestro Redis en Docker
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

	// Video
	videoRepo := video.NewVideoRepository(db)
	videoSvc := video.NewVideoService(videoRepo, asynqClient)
	videoController := video.NewVideoController(videoSvc)

	// Vote <-- AÑADIMOS LOS COMPONENTES DE VOTE
	voteRepo := vote.NewVoteRepository(db)
	voteSvc := vote.NewVoteService(voteRepo, db)
	voteController := vote.NewVoteController(voteSvc)

	router := gin.Default()
	apiV1 := router.Group("/api/v1")
	{
		user.SignUpUserRoutes(apiV1, userController)

		video.SignUpVideoRoutes(apiV1, videoController, authMiddleware)

		vote.SignUpVoteRoutes(apiV1, voteController, authMiddleware)
	}

	log.Println("Server is running on port 9090")
	router.Run(":9090")
}
