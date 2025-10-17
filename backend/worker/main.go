package main

import (
	"anb-app/src/video" // Importamos el paquete de video
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hibiken/asynq"
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

// Función de conexión directa a PostgreSQL hardcodeada
func connectPostgreSQL() *gorm.DB {
	// Configuración hardcodeada directa
	dbHost := "anb-app-db.cd6qswmk4njt.us-east-1.rds.amazonaws.com"
	dbPort := "5432"
	dbUser := "anb_user"
	dbPassword := "anb_password"
	dbName := "anb_db"
	dbSSLMode := "require"

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

func (p *TaskProcessor) HandleProcessVideoTask(ctx context.Context, t *asynq.Task) error {
	var payload video.VideoProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	retryCount, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	log.Printf("--- WORKER: Received task for video ID: %d (Retry: %d/%d) ---", payload.VideoID, retryCount, maxRetry)

	videoRecord, err := p.videoRepo.FindByID(payload.VideoID)
	if err != nil || videoRecord == nil {
		log.Printf("ERROR: Video %d not found, skipping retry.", payload.VideoID)
		return asynq.SkipRetry
	}

	processingErr := p.processVideo(videoRecord)

	if processingErr != nil {
		log.Printf("ERROR processing video ID %d: %v", payload.VideoID, processingErr)

		if retryCount >= maxRetry-1 {
			log.Printf("Task for video ID %d has failed permanently. Updating status to 'failed'.", payload.VideoID)

			videoRecord.Status = "failed"
			if updateErr := p.videoRepo.Update(videoRecord); updateErr != nil {
				return fmt.Errorf("task failed permanently and could not update status: %w (original error: %v)", updateErr, processingErr)
			}
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

	// Conexión directa hardcodeada a PostgreSQL
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

	// Redis hardcodeado para Docker
	redisAddr := "redis-anb:6379"

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
	processor := NewTaskProcessor(db, videoRepo, s3Client, s3Bucket)

	mux := asynq.NewServeMux()
	mux.HandleFunc(video.TypeVideoProcess, processor.HandleProcessVideoTask)

	log.Println(" ANB Worker is running and connected to PostgreSQL...")
	log.Println(" Waiting for video processing tasks...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("could not run asynq server: %v", err)
	}
}
