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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDClaim.(uint)

	err = vc.voteService.CreateVote(userID, uint(videoID))
	if err != nil {
		if strings.Contains(err.Error(), "already voted") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You have already voted for this video."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote registered successfully."})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userID := userIDClaim.(uint)

	err = vc.voteService.DeleteVote(userID, uint(videoID))
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			c.JSON(http.StatusNotFound, gin.H{"error": "You have not voted for this video yet."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vote removed successfully."})
}
