package main

import (
	"anb-app/src/queue"
	"anb-app/src/video" // Importamos el paquete de video
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TaskProcessor struct {
	db         *gorm.DB
	videoRepo  video.VideoRepository
	s3Client   *s3.Client
	bucketName string
}

func NewTaskProcessor(db *gorm.DB, videoRepo video.VideoRepository, s3Client *s3.Client, bucketName string) *TaskProcessor {
	return &TaskProcessor{
		db:         db,
		videoRepo:  videoRepo,
		s3Client:   s3Client,
		bucketName: bucketName,
	}
}

// Función de conexión directa a PostgreSQL usando variables de entorno
func connectPostgreSQL() *gorm.DB {
	// Leer variables de entorno
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "anb_db"
	}
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	// Crear DSN de conexión
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSLMode)

	log.Printf("Conectando a RDS: %s:%s/%s con SSL", dbHost, dbPort, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatalf("Error fatal al conectar a la base de datos: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Error al obtener la instancia SQL DB: %v", err)
	}

	// Configurar el pool de conexiones
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verificar la conexión
	if err = sqlDB.Ping(); err != nil {
		log.Fatalf("Error al hacer ping a la base de datos: %v", err)
	}

	log.Println("Conexión a la base de datos establecida exitosamente.")
	return db
}

// Helper to download file from S3 to local path
func (p *TaskProcessor) downloadFromS3(s3Key, localPath string) error {
	result, err := p.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer result.Body.Close()

	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Helper to upload file from local path to S3
func (p *TaskProcessor) uploadToS3(localPath, s3Key string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	_, err = p.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(p.bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

func (p *TaskProcessor) processVideo(videoRecord *video.Video) error {
	log.Printf("Processing video '%s'...", videoRecord.Title)

	// video.OriginalURL is now S3 key (e.g., "originals/123.mp4")
	s3Key := videoRecord.OriginalURL
	baseName := strings.TrimSuffix(filepath.Base(s3Key), filepath.Ext(s3Key))

	// Create temp directory
	tempDir := "/tmp/video-processing"
	os.MkdirAll(tempDir, os.ModePerm)
	defer os.RemoveAll(tempDir) // Clean up at the end

	// Step 1: Download original video from S3
	log.Println("Step 1: Downloading original video from S3...")
	tempOriginalPath := filepath.Join(tempDir, baseName+"_original.mp4")
	if err := p.downloadFromS3(s3Key, tempOriginalPath); err != nil {
		return fmt.Errorf("failed to download original video: %w", err)
	}

	// Step 2: Process with FFmpeg
	log.Println("Step 2: Trimming and scaling user video...")
	introVideoPath := "/app/intro/anb.mp4"
	tempProcessedPath := filepath.Join(tempDir, baseName+"_processed.mp4")
	concatListPath := filepath.Join(tempDir, baseName+"_list.txt")
	finalOutputPath := filepath.Join(tempDir, baseName+"_final.mp4")

	cmd1 := exec.Command("ffmpeg", "-y", "-i", tempOriginalPath, "-t", "30", "-vf", "scale=1280:720,setdar=16/9", "-preset", "fast", tempProcessedPath)
	if err := runFFmpegCommand(cmd1); err != nil {
		return fmt.Errorf("ffmpeg trim/scale failed: %w", err)
	}

	// Step 3: Create concatenation list
	log.Println("Step 3: Creating concatenation list...")
	concatContent := fmt.Sprintf("file '%s'\nfile '%s'\nfile '%s'", introVideoPath, tempProcessedPath, introVideoPath)
	if err := os.WriteFile(concatListPath, []byte(concatContent), 0644); err != nil {
		return fmt.Errorf("failed to create concat list: %w", err)
	}

	// Step 4: Concatenate videos
	log.Println("Step 4: Concatenating videos...")
	cmd2 := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", concatListPath, "-c", "copy", finalOutputPath)
	if err := runFFmpegCommand(cmd2); err != nil {
		return fmt.Errorf("ffmpeg concat failed: %w", err)
	}

	// Step 5: Upload processed video to S3
	log.Println("Step 5: Uploading processed video to S3...")
	processedS3Key := fmt.Sprintf("processed/%s.mp4", baseName)
	if err := p.uploadToS3(finalOutputPath, processedS3Key); err != nil {
		return fmt.Errorf("failed to upload processed video: %w", err)
	}

	// Step 6: Update database with S3 key
	log.Println("Step 6: Updating database...")
	videoRecord.Status = "processed"
	now := time.Now()
	videoRecord.ProcessedAt = &now
	videoRecord.ProcessedURL = processedS3Key // Store S3 key
	if err := p.videoRepo.Update(videoRecord); err != nil {
		return fmt.Errorf("failed to update video record %d: %w", videoRecord.ID, err)
	}

	log.Printf("Successfully processed video ID: %d", videoRecord.ID)
	return nil
}

func (p *TaskProcessor) HandleProcessVideoTask(ctx context.Context, task *queue.Task) error {
	log.Printf("--- WORKER: Processing task for video ID: %d ---", task.Payload.VideoID)

	videoRecord, err := p.videoRepo.FindByID(task.Payload.VideoID)
	if err != nil || videoRecord == nil {
		log.Printf("ERROR: Video %d not found", task.Payload.VideoID)
		return fmt.Errorf("video not found: %w", err)
	}

	processingErr := p.processVideo(videoRecord)

	if processingErr != nil {
		log.Printf("ERROR processing video ID %d: %v", task.Payload.VideoID, processingErr)

		// Marcar como fallido en la base de datos si ya se reintentó muchas veces
		// SQS manejará los reintentos automáticamente
		videoRecord.Status = "failed"
		if updateErr := p.videoRepo.Update(videoRecord); updateErr != nil {
			return fmt.Errorf("task failed and could not update status: %w (original error: %v)", updateErr, processingErr)
		}

		return processingErr
	}

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
	log.Println("Conectando worker a PostgreSQL...")

	// Conexión a PostgreSQL
	db := connectPostgreSQL()
	log.Println("Worker conectado a PostgreSQL exitosamente")

	// Initialize S3 client
	s3Bucket := os.Getenv("S3_BUCKET_NAME")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is required")
	}
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)
	log.Printf("S3 Client initialized: bucket=%s, region=%s", s3Bucket, awsRegion)

	// Inicializar SQS Consumer
	sqsQueueURL := os.Getenv("SQS_QUEUE_URL")
	if sqsQueueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable is required")
	}

	sqsConsumer, err := queue.NewSQSConsumer(context.Background(), sqsQueueURL, awsRegion)
	if err != nil {
		log.Fatalf("Failed to initialize SQS consumer: %v", err)
	}
	defer sqsConsumer.Close()

	log.Printf("SQS Consumer initialized: queue=%s", sqsQueueURL)

	videoRepo := video.NewVideoRepository(db)
	processor := NewTaskProcessor(db, videoRepo, s3Client, s3Bucket)

	log.Println(" ANB Worker is running and connected to PostgreSQL...")
	log.Println(" Waiting for video processing tasks from SQS...")

	// Iniciar servidor HTTP para health checks en un goroutine
	go startHealthCheckServer()

	// Loop infinito para recibir y procesar mensajes de SQS
	for {
		ctx := context.Background()

		// Recibir tarea de SQS (long polling de 20 segundos)
		task, err := sqsConsumer.ReceiveTask(ctx)
		if err != nil {
			log.Printf("Error receiving task from SQS: %v", err)
			time.Sleep(5 * time.Second) // Esperar antes de reintentar
			continue
		}

		// Si no hay mensajes, continuar esperando
		if task == nil {
			continue
		}

		// Procesar la tarea
		log.Printf("Processing task: %s for video ID: %d", task.ID, task.Payload.VideoID)

		processErr := processor.HandleProcessVideoTask(ctx, task)

		if processErr != nil {
			log.Printf("Task %s failed: %v", task.ID, processErr)
			// Marcar como fallida (SQS lo reintentará)
			if err := sqsConsumer.FailTask(ctx, task); err != nil {
				log.Printf("Error marking task as failed: %v", err)
			}
		} else {
			log.Printf("Task %s completed successfully", task.ID)
			// Eliminar el mensaje de SQS
			if err := sqsConsumer.CompleteTask(ctx, task); err != nil {
				log.Printf("Error completing task: %v", err)
			}
		}
	}
}

// startHealthCheckServer inicia un servidor HTTP simple para health checks del ALB
func startHealthCheckServer() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("HEALTH_CHECK_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Health check server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start health check server: %v", err)
	}
}
