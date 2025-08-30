package main

import (
	"log"
	"project-one/src/auth"
	"project-one/src/user"
	"project-one/src/video"
	"project-one/src/vote"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("anb.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	db.AutoMigrate(&user.User{}, &video.Video{}, &vote.Vote{})

	jwtSecret := "MI_CLAVE_SECRETA_SUPREMAMENTE_SEGURA"

	// Auth
	authSvc := auth.NewAuthService(jwtSecret)
	authMiddleware := authSvc.AuthMiddleware()

	// User
	userRepo := user.NewUserRepository(db)
	userSvc := user.NewUserService(userRepo, authSvc)
	userController := user.NewUserController(userSvc)

	// Video
	videoRepo := video.NewVideoRepository(db)
	videoSvc := video.NewVideoService(videoRepo)
	videoController := video.NewVideoController(videoSvc)

	// Vote <-- AÑADIMOS LOS COMPONENTES DE VOTE
	voteRepo := vote.NewVoteRepository(db)
	voteSvc := vote.NewVoteService(voteRepo, db)
	voteController := vote.NewVoteController(voteSvc)

	router := gin.Default()
	apiV1 := router.Group("/api/v1")
	{
		user.RegisterUserRoutes(apiV1, userController)

		video.RegisterVideoRoutes(apiV1, videoController, authMiddleware)

		vote.RegisterVoteRoutes(apiV1, voteController, authMiddleware)
	}

	log.Println("Server is running on port 8080")
	router.Run(":8080")
}
