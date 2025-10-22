package video

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type VideoService interface {
	Upload(ctx *gin.Context, req *UploadVideoRequest, file *multipart.FileHeader, userID uint) (*VideoResponse, error)
	ListByUserID(userID uint) ([]VideoResponse, error)
	GetByID(videoID uint, userID uint) (*VideoResponse, error)
	Delete(videoID uint, userID uint) error
	ListPublic() ([]VideoResponse, error)
	MarkAsProcessed(videoID uint, userID uint) (*VideoResponse, error)
	GetRankings() ([]RankingResponse, error)
}

type VideoController struct {
	videoService VideoService
	validate     *validator.Validate
}

func NewVideoController(videoService VideoService) *VideoController {
	return &VideoController{
		videoService: videoService,
		validate:     validator.New(),
	}
}

func (vc *VideoController) Upload(c *gin.Context) {
	userIDClaim, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, ok := userIDClaim.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in token"})
		return
	}

	req := new(UploadVideoRequest)
	if err := c.ShouldBind(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}
	if err := vc.validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Video file is required"})
		return
	}

	videoResponse, err := vc.videoService.Upload(c, req, file, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Video subido correctamente. Procesamiento en curso.",
		"task_id": fmt.Sprintf("%d-%d", videoResponse.ID, time.Now().Unix()),
	})
}

func (vc *VideoController) ListMyVideos(c *gin.Context) {
	userIDClaim, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID, ok := userIDClaim.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format in token"})
		return
	}

	videos, err := vc.videoService.ListByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve videos"})
		return
	}

	c.JSON(http.StatusOK, videos)
}

func (vc *VideoController) GetVideoByID(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}

	userIDClaim, _ := c.Get("userID")
	userID := userIDClaim.(uint)

	video, err := vc.videoService.GetByID(uint(videoID), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "The video with the specified ID does not exist or does not belong to the user."}) // <-- English
			return
		}
		if strings.Contains(err.Error(), "permission") {
			c.JSON(http.StatusForbidden, gin.H{"error": "The authenticated user does not have permission to access this video."}) // <-- English
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, video)
}

// Download endpoint no longer needed - videos are downloaded directly from S3 using presigned URLs
// The presigned URLs in OriginalURL and ProcessedURL can be used directly for downloads
func (vc *VideoController) Download(c *gin.Context) {
	c.JSON(http.StatusGone, gin.H{
		"error":   "This endpoint is deprecated",
		"message": "Videos are now served directly from S3. Use the OriginalURL or ProcessedURL from the video details.",
	})
}

func (vc *VideoController) DeleteVideo(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}

	userIDClaim, _ := c.Get("userID")
	userID := userIDClaim.(uint)

	err = vc.videoService.Delete(uint(videoID), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Video with id video_id does not exist or don't belongs to user."})
			return
		}
		if strings.Contains(err.Error(), "permission") {
			c.JSON(http.StatusForbidden, gin.H{"error": "User authenticated does not have permitions on this video."})
			return
		}
		if strings.Contains(err.Error(), "cannot delete") {
			// 400 Bad Request
			c.JSON(http.StatusBadRequest, gin.H{"error": "Video can not be deleted because conditions are'nt matched."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Éxito: 200 OK con el mensaje y video_id
	c.JSON(http.StatusOK, gin.H{
		"message":  "El video ha sido eliminado exitosamente.",
		"video_id": videoID,
	})
}

func (vc *VideoController) ListPublicVideos(c *gin.Context) {
	videos, err := vc.videoService.ListPublic()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Videos can't be fetched"})
		return
	}

	c.JSON(http.StatusOK, videos)
}

func (vc *VideoController) MarkVideoAsProcessed(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}
	userIDClaim, _ := c.Get("userID")
	userID := userIDClaim.(uint)

	video, err := vc.videoService.MarkAsProcessed(uint(videoID), userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "permission") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, video)
}

func (vc *VideoController) GetRankings(c *gin.Context) {
	rankings, err := vc.videoService.GetRankings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve rankings."})
		return
	}

	c.JSON(http.StatusOK, rankings)
}
