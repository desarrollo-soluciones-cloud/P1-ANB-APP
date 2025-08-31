package user

import "github.com/gin-gonic/gin"

func SignUpUserRoutes(router *gin.RouterGroup, userController *UserController) {
	userRoutes := router.Group("/auth")
	{
		userRoutes.POST("/signup", userController.SignUp)

		userRoutes.POST("/login", userController.Login)
	}
	publicRoutes := router.Group("/public")
	{
		publicRoutes.GET("/rankings", userController.GetRankings)
	}
}
