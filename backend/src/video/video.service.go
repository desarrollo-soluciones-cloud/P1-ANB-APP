package video

import (
	"anb-app/src/storage"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
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
	storageSvc  storage.StorageService
}

func NewVideoService(videoRepo VideoRepository, asynqClient *asynq.Client, redisClient *redis.Client, storageSvc storage.StorageService) VideoService {
	return &videoService{
		videoRepo:   videoRepo,
		asynqClient: asynqClient,
		redisClient: redisClient,
		storageSvc:  storageSvc,
	}
}

// Helper to convert S3 key to presigned URL
func (s *videoService) getPresignedURL(s3Key string) string {
	if s3Key == "" {
		return ""
	}
	// Generate presigned URL valid for 1 hour
	url, err := s.storageSvc.GetPresignedURL(s3Key, 1*time.Hour)
	if err != nil {
		log.Printf("Error generating presigned URL for %s: %v", s3Key, err)
		return ""
	}
	return url
}

func (s *videoService) Upload(ctx *gin.Context, req *UploadVideoRequest, fileHeader *multipart.FileHeader, userID uint) (*VideoResponse, error) {
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := fmt.Sprintf("%d-%d%s", time.Now().UnixNano(), userID, ext)

	// S3 key (path in bucket)
	s3Key := fmt.Sprintf("originals/%s", newFileName)

	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Upload to S3
	if err := s.storageSvc.Upload(file, s3Key); err != nil {
		return nil, err
	}

	// Store S3 key in database
	newVideo := &Video{
		UserID:      userID,
		Title:       req.Title,
		Status:      "uploaded",
		OriginalURL: s3Key, // Store S3 key
		UploadedAt:  time.Now(),
	}

	createdVideo, err := s.videoRepo.Create(newVideo)
	if err != nil {
		// Try to cleanup S3 object if DB insert fails
		s.storageSvc.Delete(s3Key)
		return nil, err
	}

	payload, err := json.Marshal(VideoProcessPayload{VideoID: createdVideo.ID})
	if err != nil {
		return nil, err
	}

	task := asynq.NewTask(
		TypeVideoProcess,
		payload,
		asynq.MaxRetry(5),
		asynq.Timeout(10*time.Minute),
	)

	taskInfo, err := s.asynqClient.Enqueue(task)
	if err != nil {

		return nil, err
	}
	log.Printf("---> Enqueued task to process video ID: %d, Task ID: %s", createdVideo.ID, taskInfo.ID)

	// Generate presigned URLs for response
	originalPresignedURL := s.getPresignedURL(createdVideo.OriginalURL)
	processedPresignedURL := s.getPresignedURL(createdVideo.ProcessedURL)

	response := &VideoResponse{
		ID:           createdVideo.ID,
		UserID:       createdVideo.UserID,
		Title:        createdVideo.Title,
		Status:       createdVideo.Status,
		OriginalURL:  originalPresignedURL,
		VoteCount:    createdVideo.VoteCount,
		UploadedAt:   createdVideo.UploadedAt,
		ProcessedAt:  createdVideo.ProcessedAt,
		ProcessedURL: processedPresignedURL,
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
			OriginalURL:  s.getPresignedURL(video.OriginalURL),
			ProcessedURL: s.getPresignedURL(video.ProcessedURL),
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
		OriginalURL:  s.getPresignedURL(video.OriginalURL),
		ProcessedURL: s.getPresignedURL(video.ProcessedURL),
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

	// Delete from S3 (video.OriginalURL is now S3 key)
	if err := s.storageSvc.Delete(video.OriginalURL); err != nil {
		log.Printf("Warning: Failed to delete S3 object %s: %v", video.OriginalURL, err)
	}

	// Also delete processed video if exists
	if video.ProcessedURL != "" {
		if err := s.storageSvc.Delete(video.ProcessedURL); err != nil {
			log.Printf("Warning: Failed to delete S3 object %s: %v", video.ProcessedURL, err)
		}
	}

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
			OriginalURL:  s.getPresignedURL(video.OriginalURL),
			ProcessedURL: s.getPresignedURL(video.ProcessedURL),
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
	// Store S3 key for processed video
	video.ProcessedURL = fmt.Sprintf("processed/%s.mp4", baseName)

	if err := s.videoRepo.Update(video); err != nil {
		return nil, err
	}

	response := &VideoResponse{
		ID:           video.ID,
		UserID:       video.UserID,
		Title:        video.Title,
		Status:       video.Status,
		OriginalURL:  s.getPresignedURL(video.OriginalURL),
		ProcessedURL: s.getPresignedURL(video.ProcessedURL),
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
