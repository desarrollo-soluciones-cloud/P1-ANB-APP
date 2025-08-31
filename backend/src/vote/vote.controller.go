package vote

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type VoteService interface {
	CreateVote(userID uint, videoID uint) error
	DeleteVote(userID uint, videoID uint) error
}

type VoteController struct {
	voteService VoteService
}

func NewVoteController(voteService VoteService) *VoteController {
	return &VoteController{
		voteService: voteService,
	}
}

func (vc *VoteController) Create(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}
	userIDClaim, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication"}) // <-- Mensaje ajustado
		return
	}
	userID := userIDClaim.(uint)

	err = vc.voteService.CreateVote(userID, uint(videoID))
	if err != nil {
		if strings.Contains(err.Error(), "already voted") {
			// 400 Bad Request
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already voted for this video."}) // <-- Mensaje ajustado
			return
		}
		// AÑADIMOS ESTE CASO PARA EL 404
		if strings.Contains(err.Error(), "video not found") {
			// 404 Not Found
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 200 OK
	c.JSON(http.StatusOK, gin.H{"message": "Vote successfully registered."}) // <-- Mensaje ajustado
}
func (vc *VoteController) Delete(c *gin.Context) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}
	userIDClaim, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication."})
		return
	}
	userID := userIDClaim.(uint)

	err = vc.voteService.DeleteVote(userID, uint(videoID))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			// 404 Not Found
			c.JSON(http.StatusNotFound, gin.H{"error": "Vote is non existent, cannot be deleted."}) // <-- Mensaje ajustado
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 200 OK
	c.JSON(http.StatusOK, gin.H{"message": "Voto successfully deleted."}) // <-- Mensaje ajustado
}
