package video

import "github.com/gin-gonic/gin"

func RegisterVideoRoutes(router *gin.RouterGroup, videoController *VideoController, authMiddleware gin.HandlerFunc) {
	videoRoutes := router.Group("/videos", authMiddleware)
	{
		videoRoutes.POST("/upload", videoController.Upload)

		videoRoutes.GET("", videoController.ListMyVideos)

		videoRoutes.GET("/:video_id", videoController.GetVideoByID)

		videoRoutes.DELETE("/:video_id", videoController.DeleteVideo)
	}
}
