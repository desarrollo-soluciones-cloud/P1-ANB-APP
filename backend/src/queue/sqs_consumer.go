package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSConsumer implementa QueueConsumer usando Amazon SQS
type SQSConsumer struct {
	client         *sqs.Client
	queueURL       string
	receiptHandles map[string]string // Mapeo de Task ID a Receipt Handle
}

// MessagePayload estructura del mensaje recibido de SQS
type MessagePayload struct {
	TaskType  string      `json:"task_type"`
	Payload   TaskPayload `json:"payload"`
	MaxRetry  int         `json:"max_retry"`
	Timeout   float64     `json:"timeout"`
	CreatedAt int64       `json:"created_at"`
}

// NewSQSConsumer crea un nuevo consumidor de SQS
func NewSQSConsumer(ctx context.Context, queueURL string, region string) (*SQSConsumer, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

	log.Printf("SQS Consumer initialized: queue=%s, region=%s", queueURL, region)

	return &SQSConsumer{
		client:         client,
		queueURL:       queueURL,
		receiptHandles: make(map[string]string),
	}, nil
}

// ReceiveTask recibe la siguiente tarea disponible de SQS
func (s *SQSConsumer) ReceiveTask(ctx context.Context) (*Task, error) {
	// Long polling con 20 segundos (máximo de SQS)
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20, // Long polling
		MessageAttributeNames: []string{
			"All",
		},
		AttributeNames: []types.QueueAttributeName{
			"ApproximateReceiveCount",
		},
	}

	result, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive message from SQS: %w", err)
	}

	// Si no hay mensajes, retornar nil (no es un error)
	if len(result.Messages) == 0 {
		return nil, nil
	}

	message := result.Messages[0]

	// Parsear el body del mensaje
	var msgPayload MessagePayload
	if err := json.Unmarshal([]byte(aws.ToString(message.Body)), &msgPayload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message body: %w", err)
	}

	// Crear la tarea
	task := &Task{
		ID:      aws.ToString(message.MessageId),
		Type:    msgPayload.TaskType,
		Payload: msgPayload.Payload,
	}

	// Guardar el receipt handle para poder completar/fallar la tarea después
	s.receiptHandles[task.ID] = aws.ToString(message.ReceiptHandle)

	// Obtener el retry count
	retryCount := 0
	if receiveCount, ok := message.Attributes["ApproximateReceiveCount"]; ok {
		retryCount, _ = strconv.Atoi(receiveCount)
		retryCount-- // Restar 1 porque el primer intento no es un retry
	}

	log.Printf("--- WORKER: Received task for video ID: %d (Retry: %d/%d) ---",
		msgPayload.Payload.VideoID, retryCount, msgPayload.MaxRetry)

	return task, nil
}

// CompleteTask elimina el mensaje de SQS (marca como completado)
func (s *SQSConsumer) CompleteTask(ctx context.Context, task *Task) error {
	receiptHandle, ok := s.receiptHandles[task.ID]
	if !ok {
		return fmt.Errorf("receipt handle not found for task %s", task.ID)
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	_, err := s.client.DeleteMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete message from SQS: %w", err)
	}

	delete(s.receiptHandles, task.ID)
	log.Printf("Task %s completed successfully", task.ID)

	return nil
}

// FailTask devuelve el mensaje a la cola (no lo elimina, SQS lo reintentará)
func (s *SQSConsumer) FailTask(ctx context.Context, task *Task) error {
	receiptHandle, ok := s.receiptHandles[task.ID]
	if !ok {
		return fmt.Errorf("receipt handle not found for task %s", task.ID)
	}

	// Cambiar visibility timeout a 0 para que esté disponible inmediatamente
	input := &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.queueURL),
		ReceiptHandle:     aws.String(receiptHandle),
		VisibilityTimeout: 0, // Disponible inmediatamente
	}

	_, err := s.client.ChangeMessageVisibility(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to change message visibility: %w", err)
	}

	delete(s.receiptHandles, task.ID)
	log.Printf("Task %s marked as failed, will be retried", task.ID)

	return nil
}

// Close cierra el cliente
func (s *SQSConsumer) Close() error {
	return nil
}
