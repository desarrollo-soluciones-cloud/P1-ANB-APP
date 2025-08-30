package video

import "github.com/gin-gonic/gin"

func SignUpVideoRoutes(router *gin.RouterGroup, vc *VideoController, authMiddleware gin.HandlerFunc) {

	protectedRoutes := router.Group("/videos", authMiddleware)
	{
		// POST /api/v1/videos/upload
		protectedRoutes.POST("/upload", vc.Upload)

		// GET /api/v1/videos
		protectedRoutes.GET("", vc.ListMyVideos)

		// GET /api/v1/videos/:video_id
		protectedRoutes.GET("/:video_id", vc.GetVideoByID)

		// DELETE /api/v1/videos/:video_id
		protectedRoutes.DELETE("/:video_id", vc.DeleteVideo)

		// POST /api/v1/videos/:video_id/mark-processed RUTA DE PRUEBA
		protectedRoutes.POST("/:video_id/mark-processed", vc.MarkVideoAsProcessed)
	}

	publicRoutes := router.Group("/public/videos")
	{
		publicRoutes.GET("", vc.ListPublicVideos)
	}
}
