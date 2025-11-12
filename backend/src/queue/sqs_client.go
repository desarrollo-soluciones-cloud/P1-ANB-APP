package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SQSClient implementa QueueClient usando Amazon SQS
type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSClient crea un nuevo cliente de SQS
func NewSQSClient(ctx context.Context, queueURL string, region string) (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

	log.Printf("SQS Client initialized: queue=%s, region=%s", queueURL, region)

	return &SQSClient{
		client:   client,
		queueURL: queueURL,
	}, nil
}

// EnqueueTask encola una nueva tarea en SQS
func (s *SQSClient) EnqueueTask(ctx context.Context, taskType string, payload TaskPayload, maxRetry int, timeout time.Duration) (string, error) {
	// Crear el mensaje con metadata
	message := map[string]interface{}{
		"task_type":  taskType,
		"payload":    payload,
		"max_retry":  maxRetry,
		"timeout":    timeout.Seconds(),
		"created_at": time.Now().Unix(),
	}

	messageBody, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Enviar mensaje a SQS
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.queueURL),
		MessageBody: aws.String(string(messageBody)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"TaskType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(taskType),
			},
			"VideoID": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(fmt.Sprintf("%d", payload.VideoID)),
			},
		},
	}

	result, err := s.client.SendMessage(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to send message to SQS: %w", err)
	}

	messageID := aws.ToString(result.MessageId)
	log.Printf("---> Enqueued task to SQS for video ID: %d, Message ID: %s", payload.VideoID, messageID)

	return messageID, nil
}

// Close cierra el cliente (SQS no requiere cerrar conexión explícitamente)
func (s *SQSClient) Close() error {
	return nil
}
