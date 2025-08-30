package user

import "github.com/gin-gonic/gin"

func RegisterUserRoutes(router *gin.RouterGroup, userController *UserController) {
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/register", userController.Register)

		userRoutes.POST("/login", userController.Login)
	}
}
