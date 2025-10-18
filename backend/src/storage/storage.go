package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService interface {
	Upload(file multipart.File, s3Key string) error
	Delete(s3Key string) error
	GetPresignedURL(s3Key string, expiration time.Duration) (string, error)
}

type s3StorageService struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucketName    string
	region        string
}

func NewS3StorageService(bucketName, region string) (StorageService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(client)

	return &s3StorageService{
		client:        client,
		presignClient: presignClient,
		bucketName:    bucketName,
		region:        region,
	}, nil
}

func (s *s3StorageService) Upload(file multipart.File, s3Key string) error {
	ctx := context.TODO()

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
		Body:   file,
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

func (s *s3StorageService) Delete(s3Key string) error {
	ctx := context.TODO()

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from S3: %w", err)
	}

	return nil
}

func (s *s3StorageService) GetPresignedURL(s3Key string, expiration time.Duration) (string, error) {
	ctx := context.TODO()

	request, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(expiration))

	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}
