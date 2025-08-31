package main

import (
	"anb-app/src/database" // Usar PostgreSQL como el API
	"anb-app/src/video"    // Importamos el paquete de video
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type TaskProcessor struct {
	db        *gorm.DB
	videoRepo video.VideoRepository
}

func NewTaskProcessor(db *gorm.DB, videoRepo video.VideoRepository) *TaskProcessor {
	return &TaskProcessor{db: db, videoRepo: videoRepo}
}

func (p *TaskProcessor) HandleProcessVideoTask(ctx context.Context, t *asynq.Task) error {
	var payload video.VideoProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("--- WORKER: Received task to process video ID: %d ---", payload.VideoID)

	videoRecord, err := p.videoRepo.FindByID(payload.VideoID)
	if err != nil || videoRecord == nil {
		return fmt.Errorf("video %d not found or error fetching: %w", payload.VideoID, err)
	}

	log.Printf("Processing video '%s'...", videoRecord.Title)

	// --- LÓGICA DE RUTAS CORREGIDA ---

	// 1. Convertir rutas a absolutas desde el directorio de ejecución actual (backend/)
	// CORRECCIÓN: Quitamos el "../"
	introVideoPath, _ := filepath.Abs("intro/anb.mp4")
	originalVideoPath, _ := filepath.Abs(videoRecord.OriginalURL) // La URL ya es relativa al CWD
	baseName := strings.TrimSuffix(filepath.Base(originalVideoPath), filepath.Ext(originalVideoPath))

	tempDir, _ := filepath.Abs("uploads/temp")
	processedDir, _ := filepath.Abs("uploads/processed")
	os.MkdirAll(tempDir, os.ModePerm)
	os.MkdirAll(processedDir, os.ModePerm)

	tempProcessedPath := filepath.Join(tempDir, baseName+"_processed.mp4")
	concatListPath := filepath.Join(tempDir, baseName+"_list.txt")
	finalOutputPath := filepath.Join(processedDir, baseName+".mp4")

	// 2. Paso 1: Recortar y escalar el video del usuario
	log.Println("Step 1: Trimming and scaling user video...")
	cmd1 := exec.Command("ffmpeg", "-y", "-i", originalVideoPath, "-t", "30", "-vf", "scale=1280:720,setdar=16/9", "-preset", "fast", tempProcessedPath)
	if err := runFFmpegCommand(cmd1); err != nil {
		return fmt.Errorf("ffmpeg trim/scale failed: %w", err)
	}

	// 3. Paso 2: Crear el archivo de lista para concatenar
	log.Println("Step 2: Creating concatenation list...")
	safeIntroPath := strings.Replace(introVideoPath, `\`, `\\`, -1)
	safeTempPath := strings.Replace(tempProcessedPath, `\`, `\\`, -1)
	concatContent := fmt.Sprintf("file '%s'\nfile '%s'\nfile '%s'", safeIntroPath, safeTempPath, safeIntroPath)
	if err := os.WriteFile(concatListPath, []byte(concatContent), 0644); err != nil {
		return fmt.Errorf("failed to create concat list: %w", err)
	}

	// 4. Paso 3: Concatenar los videos
	log.Println("Step 3: Concatenating videos...")
	cmd2 := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", concatListPath, "-c", "copy", finalOutputPath)
	if err := runFFmpegCommand(cmd2); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	// 5. Limpieza de archivos temporales
	log.Println("Step 4: Cleaning up temporary files...")
	os.Remove(tempProcessedPath)
	os.Remove(concatListPath)

	// 6. Actualizar el registro en la base de datos
	videoRecord.Status = "processed"
	now := time.Now()
	videoRecord.ProcessedAt = &now
	videoRecord.ProcessedURL = fmt.Sprintf("uploads/processed/%s.mp4", baseName) // Guardar ruta relativa simple
	if err := p.videoRepo.Update(videoRecord); err != nil {
		return fmt.Errorf("failed to update video record %d: %w", payload.VideoID, err)
	}

	log.Printf("--- WORKER: Finished processing video ID: %d ---", payload.VideoID)
	return nil
}

func runFFmpegCommand(cmd *exec.Cmd) error {
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("FFmpeg command failed: %s\n", err)
		log.Printf("FFmpeg stderr: %s\n", stderr.String())
		return err
	}
	return nil
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Advertencia: No se pudo cargar .env, usando valores por defecto")
	}

	log.Println("Conectando worker a PostgreSQL...")
	db := database.ConnectDB()
	log.Println("Worker conectado a PostgreSQL exitosamente")

	redisAddr := os.Getenv("REDIS_ADDR")

	log.Printf("Configurando Asynq con Redis en: %s", redisAddr)
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Queues: map[string]int{
				"default": 10,
			},
		},
	)

	videoRepo := video.NewVideoRepository(db)
	processor := NewTaskProcessor(db, videoRepo)

	mux := asynq.NewServeMux()
	mux.HandleFunc(video.TypeVideoProcess, processor.HandleProcessVideoTask)

	log.Println(" ANB Worker is running and connected to PostgreSQL...")
	log.Println(" Waiting for video processing tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run asynq server: %v", err)
	}
}
