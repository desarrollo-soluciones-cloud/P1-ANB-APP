package video

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

const (
	TypeVideoProcess = "task:video:process"
)

type VideoProcessPayload struct {
	VideoID uint
}

type VideoRepository interface {
	Create(video *Video) (*Video, error)
	FindByUserID(userID uint) ([]Video, error)
	FindByID(videoID uint) (*Video, error)
	Delete(videoID uint) error
	FindPublic() ([]Video, error)
	Update(video *Video) error
	GetRankings() ([]RankingResponse, error)
}

type videoService struct {
	videoRepo   VideoRepository
	asynqClient *asynq.Client
	redisClient *redis.Client
}

func NewVideoService(videoRepo VideoRepository, asynqClient *asynq.Client, redisClient *redis.Client) VideoService {
	return &videoService{
		videoRepo:   videoRepo,
		asynqClient: asynqClient,
		redisClient: redisClient,
	}
}

func (s *videoService) Upload(ctx *gin.Context, req *UploadVideoRequest, file *multipart.FileHeader, userID uint) (*VideoResponse, error) {
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), userID, ext)

	uploadPath := "./uploads/originals"
	if err := os.MkdirAll(uploadPath, os.ModePerm); err != nil {
		return nil, err
	}
	filePath := filepath.Join(uploadPath, newFileName)

	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		return nil, err
	}

	newVideo := &Video{
		UserID:      userID,
		Title:       req.Title,
		Status:      "uploaded",
		OriginalURL: filePath,
		UploadedAt:  time.Now(),
	}

	createdVideo, err := s.videoRepo.Create(newVideo)
	if err != nil {
		os.Remove(filePath)
		return nil, err
	}

	payload, err := json.Marshal(VideoProcessPayload{VideoID: createdVideo.ID})
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(TypeVideoProcess, payload)

	taskInfo, err := s.asynqClient.Enqueue(task)
	if err != nil {

		return nil, err
	}
	log.Printf("---> Enqueued task to process video ID: %d, Task ID: %s", createdVideo.ID, taskInfo.ID)

	response := &VideoResponse{
		ID:           createdVideo.ID,
		UserID:       createdVideo.UserID,
		Title:        createdVideo.Title,
		Status:       createdVideo.Status,
		OriginalURL:  createdVideo.OriginalURL,
		VoteCount:    createdVideo.VoteCount,
		UploadedAt:   createdVideo.UploadedAt,
		ProcessedAt:  createdVideo.ProcessedAt,
		ProcessedURL: createdVideo.ProcessedURL,
	}

	return response, nil
}

func (s *videoService) ListByUserID(userID uint) ([]VideoResponse, error) {
	videos, err := s.videoRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	var videoResponses []VideoResponse
	for _, video := range videos {
		response := VideoResponse{
			ID:           video.ID,
			UserID:       video.UserID,
			Title:        video.Title,
			Status:       video.Status,
			OriginalURL:  video.OriginalURL,
			ProcessedURL: video.ProcessedURL,
			VoteCount:    video.VoteCount,
			UploadedAt:   video.UploadedAt,
			ProcessedAt:  video.ProcessedAt,
		}
		videoResponses = append(videoResponses, response)
	}

	return videoResponses, nil
}

func (s *videoService) GetByID(videoID uint, userID uint) (*VideoResponse, error) {
	video, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, errors.New("video not found")
	}

	if video.UserID != userID {
		return nil, errors.New("user does not have permission to access this video")
	}

	response := &VideoResponse{
		ID:           video.ID,
		UserID:       video.UserID,
		Title:        video.Title,
		Status:       video.Status,
		OriginalURL:  video.OriginalURL,
		ProcessedURL: video.ProcessedURL,
		VoteCount:    video.VoteCount,
		UploadedAt:   video.UploadedAt,
		ProcessedAt:  video.ProcessedAt,
	}

	return response, nil
}

func (s *videoService) Delete(videoID uint, userID uint) error {
	video, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return err
	}
	if video == nil {
		return errors.New("video not found")
	}

	if video.UserID != userID {
		return errors.New("user does not have permission to delete this video")
	}

	if video.Status != "uploaded" {
		return errors.New("cannot delete a video that has been processed or published")
	}

	err = s.videoRepo.Delete(videoID)
	if err != nil {
		return err
	}

	_ = os.Remove(video.OriginalURL)

	return nil
}

func (s *videoService) ListPublic() ([]VideoResponse, error) {
	videos, err := s.videoRepo.FindPublic()
	if err != nil {
		return nil, err
	}

	var videoResponses []VideoResponse
	for _, video := range videos {
		response := VideoResponse{
			ID:           video.ID,
			UserID:       video.UserID,
			Title:        video.Title,
			Status:       video.Status,
			OriginalURL:  video.OriginalURL,
			ProcessedURL: video.ProcessedURL,
			VoteCount:    video.VoteCount,
			UploadedAt:   video.UploadedAt,
			ProcessedAt:  video.ProcessedAt,
		}
		videoResponses = append(videoResponses, response)
	}

	return videoResponses, nil
}

func (s *videoService) MarkAsProcessed(videoID uint, userID uint) (*VideoResponse, error) {
	video, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, errors.New("video not found")
	}

	if video.UserID != userID {
		return nil, errors.New("user does not have permission to modify this video")
	}

	video.Status = "processed"
	now := time.Now()
	video.ProcessedAt = &now
	baseName := strings.TrimSuffix(filepath.Base(video.OriginalURL), filepath.Ext(video.OriginalURL))
	video.ProcessedURL = fmt.Sprintf("./uploads/processed/%s.mp4", baseName)

	if err := s.videoRepo.Update(video); err != nil {
		return nil, err
	}

	response := &VideoResponse{
		ID:           video.ID,
		UserID:       video.UserID,
		Title:        video.Title,
		Status:       video.Status,
		OriginalURL:  video.OriginalURL,
		ProcessedURL: video.ProcessedURL,
		VoteCount:    video.VoteCount,
		UploadedAt:   video.UploadedAt,
		ProcessedAt:  video.ProcessedAt,
	}

	return response, nil
}

func (s *videoService) GetRankings() ([]RankingResponse, error) {
	cacheKey := "rankings:videos" // Nueva clave de caché

	cachedRankings, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var rankings []RankingResponse
		json.Unmarshal([]byte(cachedRankings), &rankings)
		return rankings, nil
	}

	if err != redis.Nil {
		return nil, err
	}

	rankings, err := s.videoRepo.GetRankings()
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(rankings)
	if err != nil {
		return nil, err
	}
	s.redisClient.Set(ctx, cacheKey, jsonData, 2*time.Minute)

	return rankings, nil
}
