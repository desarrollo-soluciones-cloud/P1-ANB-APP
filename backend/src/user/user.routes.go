package user

import "github.com/gin-gonic/gin"

func RegisterUserRoutes(router *gin.RouterGroup, userController *UserController) {
	userRoutes := router.Group("/auth")
	{
		userRoutes.POST("/signup", userController.Register)

		userRoutes.POST("/login", userController.Login)
	}
}
