package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"anb-app/src/video" // Importamos el paquete de video para usar su lógica y modelos

	"github.com/glebarez/sqlite"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// TaskProcessor es una estructura para manejar las dependencias de nuestros handlers.
type TaskProcessor struct {
	db        *gorm.DB
	videoRepo video.VideoRepository
}

// NewTaskProcessor crea una nueva instancia de TaskProcessor.
func NewTaskProcessor(db *gorm.DB, videoRepo video.VideoRepository) *TaskProcessor {
	return &TaskProcessor{db: db, videoRepo: videoRepo}
}

// HandleProcessVideoTask es la función que maneja la tarea de procesamiento de video.
func (p *TaskProcessor) HandleProcessVideoTask(ctx context.Context, t *asynq.Task) error {
	var payload video.VideoProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("--- WORKER: Received task to process video ID: %d ---", payload.VideoID)

	// 1. Obtener el video de la base de datos.
	videoRecord, err := p.videoRepo.FindByID(payload.VideoID)
	if err != nil {
		return fmt.Errorf("failed to find video %d: %w", payload.VideoID, err)
	}
	if videoRecord == nil {
		return fmt.Errorf("video %d not found", payload.VideoID)
	}

	// 2. Simular el trabajo pesado de procesamiento de video.
	log.Printf("Processing video '%s'...", videoRecord.Title)
	time.Sleep(20 * time.Second) // Simula una tarea larga

	// 3. Actualizar el estado y los metadatos del video.
	videoRecord.Status = "processed"
	now := time.Now()
	videoRecord.ProcessedAt = &now
	baseName := strings.TrimSuffix(filepath.Base(videoRecord.OriginalURL), filepath.Ext(videoRecord.OriginalURL))
	videoRecord.ProcessedURL = fmt.Sprintf("./uploads/processed/%s.mp4", baseName)

	// 4. Guardar los cambios en la base de datos.
	if err := p.videoRepo.Update(videoRecord); err != nil {
		return fmt.Errorf("failed to update video record %d: %w", payload.VideoID, err)
	}

	log.Printf("--- WORKER: Finished processing video ID: %d ---", payload.VideoID)
	return nil
}

func main() {
	// --- Conexión a la Base de Datos ---
	// El worker necesita su propia conexión a la DB para actualizar los registros.
	db, err := gorm.Open(sqlite.Open("../anb.db"), &gorm.Config{}) // Ojo con la ruta "../"
	if err != nil {
		log.Fatalf("worker failed to connect database: %v", err)
	}

	// --- Configuración del Servidor de Asynq (El Consumidor) ---
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: "localhost:6379"},
		asynq.Config{
			Queues: map[string]int{
				"default": 10, // Podemos tener colas con diferentes prioridades
			},
		},
	)

	// --- Inyección de Dependencias del Worker ---
	videoRepo := video.NewVideoRepository(db)
	processor := NewTaskProcessor(db, videoRepo)

	// --- Registrar los Handlers de Tareas ---
	mux := asynq.NewServeMux()
	mux.HandleFunc(video.TypeVideoProcess, processor.HandleProcessVideoTask)

	log.Println("Worker is running...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run asynq server: %v", err)
	}
}
