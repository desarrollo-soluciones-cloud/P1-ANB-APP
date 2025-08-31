package video

import (
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
)

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
}

type videoService struct {
	videoRepo   VideoRepository
	asynqClient *asynq.Client
}

func NewVideoService(videoRepo VideoRepository, asynqClient *asynq.Client) VideoService {
	return &videoService{
		videoRepo:   videoRepo,
		asynqClient: asynqClient,
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

	// --- LÓGICA PARA ENCOLAR LA TAREA ---
	// 1. Crear el payload de la tarea.
	payload, err := json.Marshal(VideoProcessPayload{VideoID: createdVideo.ID})
	if err != nil {
		return nil, err
	}

	// 2. Crear una nueva tarea de Asynq.
	task := asynq.NewTask(TypeVideoProcess, payload)

	// 3. Encolar la tarea en Redis.
	taskInfo, err := s.asynqClient.Enqueue(task)
	if err != nil {
		// Si encolar la tarea falla, podríamos querer deshacer la subida,
		// pero por ahora solo registraremos el error.
		return nil, err
	}
	log.Printf("---> Enqueued task to process video ID: %d, Task ID: %s", createdVideo.ID, taskInfo.ID)
	// --- FIN DE LA LÓGICA DE LA TAREA ---

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
	// 1. Buscar el video para asegurarse de que existe y es del usuario.
	video, err := s.videoRepo.FindByID(videoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, errors.New("video not found")
	}

	// 2. Autorización: verificar que el video pertenece al usuario.
	if video.UserID != userID {
		return nil, errors.New("user does not have permission to modify this video")
	}

	// 3. Actualizar los campos del video.
	video.Status = "processed"
	now := time.Now()
	video.ProcessedAt = &now
	// Creamos una URL de prueba para el video procesado.
	baseName := strings.TrimSuffix(filepath.Base(video.OriginalURL), filepath.Ext(video.OriginalURL))
	video.ProcessedURL = fmt.Sprintf("./uploads/processed/%s.mp4", baseName)

	// 4. Guardar los cambios en la base de datos.
	if err := s.videoRepo.Update(video); err != nil {
		return nil, err
	}

	// 5. Mapear y devolver la entidad actualizada.
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
