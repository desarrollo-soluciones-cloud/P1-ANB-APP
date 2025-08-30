package vote

import "github.com/gin-gonic/gin"

func RegisterVoteRoutes(router *gin.RouterGroup, voteController *VoteController, authMiddleware gin.HandlerFunc) {

	voteRoutes := router.Group("/public/videos/:video_id/vote", authMiddleware)
	{
		voteRoutes.POST("", voteController.Create)

		voteRoutes.DELETE("", voteController.Delete)
	}
}
