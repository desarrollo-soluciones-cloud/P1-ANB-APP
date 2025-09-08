package video

import "github.com/gin-gonic/gin"

func SignUpVideoRoutes(router *gin.RouterGroup, vc *VideoController, authMiddleware gin.HandlerFunc) {

	protectedRoutes := router.Group("/videos", authMiddleware)
	{
		protectedRoutes.POST("/upload", vc.Upload)

		protectedRoutes.GET("", vc.ListMyVideos)

		protectedRoutes.GET("/:video_id", vc.GetVideoByID)

		// Download endpoint that streams the video file as an attachment
		protectedRoutes.GET("/:video_id/download", vc.Download)

		protectedRoutes.DELETE("/:video_id", vc.DeleteVideo)

		protectedRoutes.POST("/:video_id/mark-processed", vc.MarkVideoAsProcessed)
	}

	publicRoutes := router.Group("/public")
	{
		publicRoutes.GET("/videos", vc.ListPublicVideos)

		publicRoutes.GET("/rankings", vc.GetRankings)
	}
}
